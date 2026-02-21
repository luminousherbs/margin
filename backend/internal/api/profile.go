package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"margin.at/internal/config"
	"margin.at/internal/db"
	"margin.at/internal/xrpc"
)

type UpdateProfileRequest struct {
	DisplayName string        `json:"displayName"`
	Avatar      *xrpc.BlobRef `json:"avatar"`
	Bio         string        `json:"bio"`
	Website     string        `json:"website"`
	Links       []string      `json:"links"`
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	record := &xrpc.MarginProfileRecord{
		Type:        xrpc.CollectionProfile,
		DisplayName: req.DisplayName,
		Avatar:      req.Avatar,
		Bio:         req.Bio,
		Website:     req.Website,
		Links:       req.Links,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	if err := record.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var pdsURL string
	err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		pdsURL = client.PDS
		_, err := client.PutRecord(r.Context(), did, xrpc.CollectionProfile, "self", record)
		return err
	})

	if err != nil {
		HandleAPIError(w, r, err, "Failed to update profile: ", http.StatusInternalServerError)
		return
	}

	var avatarURL *string
	if req.Avatar != nil && req.Avatar.Ref.Link != "" {
		url := fmt.Sprintf("%s/xrpc/com.atproto.sync.getBlob?did=%s&cid=%s",
			pdsURL, session.DID, req.Avatar.Ref.Link)
		avatarURL = &url
	}

	linksJSON, _ := json.Marshal(req.Links)
	profile := &db.Profile{
		URI:         fmt.Sprintf("at://%s/%s/self", session.DID, xrpc.CollectionProfile),
		AuthorDID:   session.DID,
		DisplayName: stringPtr(req.DisplayName),
		Avatar:      avatarURL,
		Bio:         stringPtr(req.Bio),
		Website:     stringPtr(req.Website),
		LinksJSON:   stringPtr(string(linksJSON)),
		CreatedAt:   time.Now(),
		IndexedAt:   time.Now(),
	}
	h.db.UpsertProfile(profile)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}

	if did == "" {
		http.Error(w, "DID required", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(did, "did:") {
		var resolvedDID string
		err := h.db.QueryRow("SELECT did FROM sessions WHERE handle = $1 LIMIT 1", did).Scan(&resolvedDID)
		if err == nil {
			did = resolvedDID
		} else {
			resolvedDID, err = xrpc.ResolveHandle(did)
			if err == nil {
				did = resolvedDID
			}
		}
	}

	profile, err := h.db.GetProfile(did)
	if err != nil {
		http.Error(w, "Failed to fetch profile", http.StatusInternalServerError)
		return
	}

	if profile == nil {
		w.Header().Set("Content-Type", "application/json")
		if did != "" && strings.HasPrefix(did, "did:") {
			json.NewEncoder(w).Encode(map[string]string{"did": did})
		} else {
			w.Write([]byte("{}"))
		}
		return
	}

	resp := struct {
		URI         string   `json:"uri"`
		DID         string   `json:"did"`
		Handle      string   `json:"handle,omitempty"`
		DisplayName string   `json:"displayName,omitempty"`
		Avatar      string   `json:"avatar,omitempty"`
		Description string   `json:"description,omitempty"`
		Website     string   `json:"website,omitempty"`
		Links       []string `json:"links"`
		CreatedAt   string   `json:"createdAt"`
		IndexedAt   string   `json:"indexedAt"`
		Labels      []struct {
			Val string `json:"val"`
			Src string `json:"src"`
		} `json:"labels,omitempty"`
		Viewer *struct {
			Blocking  bool `json:"blocking"`
			Muting    bool `json:"muting"`
			BlockedBy bool `json:"blockedBy"`
		} `json:"viewer,omitempty"`
	}{
		URI:       profile.URI,
		DID:       profile.AuthorDID,
		CreatedAt: profile.CreatedAt.Format(time.RFC3339),
		IndexedAt: profile.IndexedAt.Format(time.RFC3339),
	}

	var handle string
	if err := h.db.QueryRow("SELECT handle FROM sessions WHERE did = $1 LIMIT 1", profile.AuthorDID).Scan(&handle); err == nil {
		resp.Handle = handle
	}

	if profile.DisplayName != nil {
		resp.DisplayName = *profile.DisplayName
	}
	if profile.Avatar != nil {
		resp.Avatar = *profile.Avatar
	}
	if profile.Bio != nil {
		resp.Description = *profile.Bio
	}
	if profile.Website != nil {
		resp.Website = *profile.Website
	}
	if profile.LinksJSON != nil && *profile.LinksJSON != "" {
		_ = json.Unmarshal([]byte(*profile.LinksJSON), &resp.Links)
	}
	if resp.Links == nil {
		resp.Links = []string{}
	}

	viewerDID := h.getViewerDID(r)
	if viewerDID != "" && viewerDID != profile.AuthorDID {
		blocking, muting, blockedBy, err := h.db.GetViewerRelationship(viewerDID, profile.AuthorDID)
		if err == nil {
			resp.Viewer = &struct {
				Blocking  bool `json:"blocking"`
				Muting    bool `json:"muting"`
				BlockedBy bool `json:"blockedBy"`
			}{
				Blocking:  blocking,
				Muting:    muting,
				BlockedBy: blockedBy,
			}
		}
	}

	subscribedLabelers := getSubscribedLabelers(h.db, viewerDID)
	if subscribedLabelers == nil {
		serviceDID := config.Get().ServiceDID
		if serviceDID != "" {
			subscribedLabelers = []string{serviceDID}
		}
	}
	if didLabels, err := h.db.GetContentLabelsForDIDs([]string{profile.AuthorDID}, subscribedLabelers); err == nil {
		if labels, ok := didLabels[profile.AuthorDID]; ok {
			for _, l := range labels {
				resp.Labels = append(resp.Labels, struct {
					Val string `json:"val"`
					Src string `json:"src"`
				}{Val: l.Val, Src: l.Src})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Failed to read avatar file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		http.Error(w, "Invalid image type. Must be JPEG or PNG.", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	var blobRef *xrpc.BlobRef
	err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var uploadErr error
		blobRef, uploadErr = client.UploadBlob(r.Context(), data, contentType)
		return uploadErr
	})

	if err != nil {
		http.Error(w, "Failed to upload avatar: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blob": blobRef,
	})
}
