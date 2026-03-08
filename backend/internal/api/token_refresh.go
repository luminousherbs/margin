package api

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"margin.at/internal/db"
	"margin.at/internal/logger"
	"margin.at/internal/oauth"
	"margin.at/internal/xrpc"
)

var ErrSessionInvalid = errors.New("session invalid")

type TokenRefresher struct {
	db         *db.DB
	privateKey *ecdsa.PrivateKey
	baseURL    string
}

func NewTokenRefresher(database *db.DB, privateKey *ecdsa.PrivateKey) *TokenRefresher {
	return &TokenRefresher{
		db:         database,
		privateKey: privateKey,
		baseURL:    os.Getenv("BASE_URL"),
	}
}

func (tr *TokenRefresher) getOAuthClient(r *http.Request) *oauth.Client {
	baseURL := tr.baseURL
	if baseURL == "" {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}

	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	clientID := baseURL + "/client-metadata.json"
	redirectURI := baseURL + "/auth/callback"

	return oauth.NewClient(clientID, redirectURI, tr.privateKey)
}

type SessionData struct {
	ID           string
	DID          string
	Handle       string
	AccessToken  string
	RefreshToken string
	DPoPKey      *ecdsa.PrivateKey
	PDS          string
}

func (tr *TokenRefresher) GetSessionWithAutoRefresh(r *http.Request) (*SessionData, error) {
	sessionID := ""

	cookie, err := r.Cookie("margin_session")
	if err == nil {
		sessionID = cookie.Value
	} else {
		sessionID = r.Header.Get("X-Session-Token")
	}

	if sessionID == "" {
		return nil, fmt.Errorf("not authenticated")
	}

	did, handle, accessToken, refreshToken, dpopKeyStr, err := tr.db.GetSession(sessionID)
	if err != nil {
		tr.db.DeleteSession(sessionID)
		return nil, fmt.Errorf("%w: session expired", ErrSessionInvalid)
	}

	block, _ := pem.Decode([]byte(dpopKeyStr))
	if block == nil {
		tr.db.DeleteSession(sessionID)
		return nil, fmt.Errorf("%w: invalid DPoP key", ErrSessionInvalid)
	}
	dpopKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		tr.db.DeleteSession(sessionID)
		return nil, fmt.Errorf("%w: invalid DPoP key", ErrSessionInvalid)
	}

	pds, err := xrpc.ResolveDIDToPDS(did)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS")
	}

	return &SessionData{
		ID:           sessionID,
		DID:          did,
		Handle:       handle,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DPoPKey:      dpopKey,
		PDS:          pds,
	}, nil
}

func (tr *TokenRefresher) RefreshSessionToken(r *http.Request, session *SessionData) (*SessionData, error) {
	if session.ID == "" {
		return nil, fmt.Errorf("invalid session ID")
	}

	oauthClient := tr.getOAuthClient(r)
	ctx := context.Background()

	meta, err := oauthClient.GetAuthServerMetadata(ctx, session.PDS)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth server metadata: %w", err)
	}

	tokenResp, _, err := oauthClient.RefreshToken(meta, session.RefreshToken, session.DPoPKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	dpopKeyBytes, err := x509.MarshalECPrivateKey(session.DPoPKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal DPoP key: %w", err)
	}
	dpopKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: dpopKeyBytes,
	})

	newRefreshToken := tokenResp.RefreshToken
	if newRefreshToken == "" {
		newRefreshToken = session.RefreshToken
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := tr.db.SaveSession(
		session.ID,
		session.DID,
		session.Handle,
		tokenResp.AccessToken,
		newRefreshToken,
		string(dpopKeyPEM),
		expiresAt,
	); err != nil {
		return nil, fmt.Errorf("failed to save refreshed session: %w", err)
	}

	logger.Info("Successfully refreshed token for user %s", session.Handle)

	return &SessionData{
		ID:           session.ID,
		DID:          session.DID,
		Handle:       session.Handle,
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: newRefreshToken,
		DPoPKey:      session.DPoPKey,
		PDS:          session.PDS,
	}, nil
}

func IsTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return bytes.Contains([]byte(errStr), []byte("invalid_token")) ||
		bytes.Contains([]byte(errStr), []byte("AuthenticationRequired")) ||
		bytes.Contains([]byte(errStr), []byte("Unauthorized")) ||
		bytes.Contains([]byte(errStr), []byte("authentication required")) ||
		bytes.Contains([]byte(errStr), []byte("TokenExpired"))
}

func (tr *TokenRefresher) ExecuteWithAutoRefresh(
	r *http.Request,
	session *SessionData,
	fn func(client *xrpc.Client, did string) error,
) error {
	client := xrpc.NewClient(session.PDS, session.AccessToken, session.DPoPKey)

	err := fn(client, session.DID)
	if err == nil {
		return nil
	}

	if !IsTokenExpiredError(err) {
		return err
	}

	logger.Info("Token expired for user %s, attempting refresh...", session.Handle)

	newSession, refreshErr := tr.RefreshSessionToken(r, session)
	if refreshErr != nil {
		logger.Error("Token refresh failed for user %s, invalidating session: %v", session.Handle, refreshErr)
		tr.db.DeleteSession(session.ID)
		return fmt.Errorf("%w: %v", ErrSessionInvalid, refreshErr)
	}

	client = xrpc.NewClient(newSession.PDS, newSession.AccessToken, newSession.DPoPKey)
	return fn(client, newSession.DID)
}

func (tr *TokenRefresher) CreateClientFromSession(session *SessionData) *xrpc.Client {
	return xrpc.NewClient(session.PDS, session.AccessToken, session.DPoPKey)
}

func HandleAPIError(w http.ResponseWriter, r *http.Request, err error, fallbackMsg string, fallbackStatus int) {
	if errors.Is(err, ErrSessionInvalid) {
		http.SetCookie(w, &http.Cookie{
			Name:     "margin_session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		})
		http.Error(w, "session expired", http.StatusUnauthorized)
		return
	}
	http.Error(w, fallbackMsg+err.Error(), fallbackStatus)
}
