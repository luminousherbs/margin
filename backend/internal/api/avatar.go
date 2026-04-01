package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) HandleAvatarProxy(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if did == "" {
		WriteBadRequest(w, "DID required")
		return
	}

	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}

	cdnURL := os.Getenv("AVATAR_CDN_URL")
	if cdnURL == "" {
		cdnURL = "https://avatars.margin.at"
	}

	w.Header().Set("Cache-Control", "public, max-age=86400")

	secret := os.Getenv("AVATAR_SHARED_SECRET")
	if secret != "" {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(did))
		sig := hex.EncodeToString(mac.Sum(nil))
		http.Redirect(w, r, fmt.Sprintf("%s/%s/%s", cdnURL, sig, did), http.StatusMovedPermanently)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s/unsigned/%s", cdnURL, did), http.StatusMovedPermanently)
}

func getProxiedAvatarURL(did, originalURL string) string {
	if originalURL == "" {
		return ""
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		return originalURL
	}

	return baseURL + "/api/avatar/" + url.PathEscape(did)
}
