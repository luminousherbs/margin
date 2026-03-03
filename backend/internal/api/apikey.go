package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"margin.at/internal/db"
	"margin.at/internal/logger"
	"margin.at/internal/xrpc"
)

type APIKeyHandler struct {
	db        *db.DB
	refresher *TokenRefresher
}

func NewAPIKeyHandler(database *db.DB, refresher *TokenRefresher) *APIKeyHandler {
	return &APIKeyHandler{db: database, refresher: refresher}
}

type CreateKeyRequest struct {
	Name string `json:"name"`
}

type CreateKeyResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"createdAt"`
}

func (h *APIKeyHandler) CreateKey(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		req.Name = "API Key"
	}

	rawKey := generateAPIKey()
	keyHash := hashAPIKey(rawKey)
	keyID := generateKeyID()

	record := xrpc.NewAPIKeyRecord(req.Name, keyHash)
	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput
	err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionAPIKey, record)
		return createErr
	})
	if err != nil {
		logger.Error("[ERROR] Failed to create API key record on PDS: %v", err)
		http.Error(w, "Failed to create key record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	cid := result.CID

	apiKey := &db.APIKey{
		ID:        keyID,
		OwnerDID:  session.DID,
		Name:      req.Name,
		KeyHash:   keyHash,
		CreatedAt: time.Now(),
		URI:       result.URI,
		CID:       &cid,
		IndexedAt: time.Now(),
	}

	if err := h.db.CreateAPIKey(apiKey); err != nil {
		logger.Error("[ERROR] Failed to insert API key into DB: %v", err)
		http.Error(w, "Failed to create key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateKeyResponse{
		ID:        keyID,
		Name:      req.Name,
		Key:       rawKey,
		CreatedAt: apiKey.CreatedAt,
	})
}

func (h *APIKeyHandler) ListKeys(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	keys, err := h.db.GetAPIKeysByOwner(session.DID)
	if err != nil {
		http.Error(w, "Failed to get keys", http.StatusInternalServerError)
		return
	}

	if keys == nil {
		keys = []db.APIKey{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"keys": keys})
}

func (h *APIKeyHandler) DeleteKey(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	keyID := chi.URLParam(r, "id")
	if keyID == "" {
		http.Error(w, "Key ID required", http.StatusBadRequest)
		return
	}

	uri, err := h.db.DeleteAPIKey(keyID, session.DID)
	if err != nil {
		http.Error(w, "Failed to delete key", http.StatusInternalServerError)
		return
	}

	if uri != "" {
		h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
			return client.DeleteRecord(r.Context(), did, xrpc.CollectionAPIKey, strings.Split(uri, "/")[len(strings.Split(uri, "/"))-1])
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

type QuickBookmarkRequest struct {
	URL         string `json:"url"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

func (h *APIKeyHandler) QuickBookmark(w http.ResponseWriter, r *http.Request) {
	apiKey, err := h.authenticateAPIKey(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req QuickBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	session, err := h.getSessionByDID(apiKey.OwnerDID)
	if err != nil {
		http.Error(w, "User session not found. Please log in to margin.at first.", http.StatusUnauthorized)
		return
	}

	urlHash := db.HashURL(req.URL)
	record := xrpc.NewBookmarkRecord(req.URL, urlHash, req.Title, req.Description)

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput
	err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionBookmark, record)
		return createErr
	})
	if err != nil {
		http.Error(w, "Failed to create bookmark: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.db.UpdateAPIKeyLastUsed(apiKey.ID)

	var titlePtr, descPtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}
	if req.Description != "" {
		descPtr = &req.Description
	}

	cid := result.CID
	bookmark := &db.Bookmark{
		URI:         result.URI,
		AuthorDID:   apiKey.OwnerDID,
		Source:      req.URL,
		SourceHash:  urlHash,
		Title:       titlePtr,
		Description: descPtr,
		CreatedAt:   time.Now(),
		IndexedAt:   time.Now(),
		CID:         &cid,
	}
	h.db.CreateBookmark(bookmark)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uri":     result.URI,
		"cid":     result.CID,
		"message": "Bookmark created successfully",
	})
}

type QuickSaveRequest struct {
	URL      string          `json:"url"`
	Text     string          `json:"text,omitempty"`
	Selector json.RawMessage `json:"selector,omitempty"`
	Color    string          `json:"color,omitempty"`
}

func (h *APIKeyHandler) QuickSave(w http.ResponseWriter, r *http.Request) {
	apiKey, err := h.authenticateAPIKey(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req QuickSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	session, err := h.getSessionByDID(apiKey.OwnerDID)
	if err != nil {
		http.Error(w, "User session not found. Please log in to margin.at first.", http.StatusUnauthorized)
		return
	}

	urlHash := db.HashURL(req.URL)

	var isHighlight bool
	if req.Selector != nil && req.Text == "" {
		isHighlight = true
	}

	var result *xrpc.CreateRecordOutput
	var createErr error

	if isHighlight {
		color := req.Color
		if color == "" {
			color = "yellow"
		}
		record := xrpc.NewHighlightRecord(req.URL, urlHash, req.Selector, color, nil)

		if err := record.Validate(); err != nil {
			http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
			result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionHighlight, record)
			return createErr
		})
		if err == nil {
			h.db.UpdateAPIKeyLastUsed(apiKey.ID)
			selectorJSON, _ := json.Marshal(req.Selector)
			selectorStr := string(selectorJSON)
			colorPtr := &color

			highlight := &db.Highlight{
				URI:          result.URI,
				AuthorDID:    apiKey.OwnerDID,
				TargetSource: req.URL,
				TargetHash:   urlHash,
				SelectorJSON: &selectorStr,
				Color:        colorPtr,
				CreatedAt:    time.Now(),
				IndexedAt:    time.Now(),
				CID:          &result.CID,
			}
			go func() {
				if err := h.db.CreateHighlight(highlight); err != nil {
					fmt.Printf("Warning: failed to index highlight in local DB: %v\n", err)
				}
			}()
		}

	} else {
		record := xrpc.NewAnnotationRecord(req.URL, urlHash, req.Text, req.Selector, "")

		if err := record.Validate(); err != nil {
			http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
			return
		}

		err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
			result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionAnnotation, record)
			return createErr
		})
		if err == nil {
			h.db.UpdateAPIKeyLastUsed(apiKey.ID)

			var selectorStrPtr *string
			if req.Selector != nil {
				b, _ := json.Marshal(req.Selector)
				s := string(b)
				selectorStrPtr = &s
			}

			bodyValue := req.Text
			var bodyValuePtr *string
			if bodyValue != "" {
				bodyValuePtr = &bodyValue
			}

			annotation := &db.Annotation{
				URI:          result.URI,
				AuthorDID:    apiKey.OwnerDID,
				Motivation:   "commenting",
				BodyValue:    bodyValuePtr,
				TargetSource: req.URL,
				TargetHash:   urlHash,
				SelectorJSON: selectorStrPtr,
				CreatedAt:    time.Now(),
				IndexedAt:    time.Now(),
				CID:          &result.CID,
			}
			go func() {
				h.db.CreateAnnotation(annotation)
			}()
		}
	}

	if err != nil {
		http.Error(w, "Failed to create record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uri":     result.URI,
		"cid":     result.CID,
		"message": "Saved successfully",
	})
}

type QuickHighlightRequest struct {
	URL      string      `json:"url"`
	Selector interface{} `json:"selector"`
	Color    string      `json:"color,omitempty"`
}

func (h *APIKeyHandler) QuickHighlight(w http.ResponseWriter, r *http.Request) {
	apiKey, err := h.authenticateAPIKey(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req QuickHighlightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" || req.Selector == nil {
		http.Error(w, "URL and selector are required", http.StatusBadRequest)
		return
	}

	session, err := h.getSessionByDID(apiKey.OwnerDID)
	if err != nil {
		http.Error(w, "User session not found. Please log in to margin.at first.", http.StatusUnauthorized)
		return
	}

	urlHash := db.HashURL(req.URL)
	color := req.Color
	if color == "" {
		color = "yellow"
	}

	record := xrpc.NewHighlightRecord(req.URL, urlHash, req.Selector, color, nil)

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput
	err = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionHighlight, record)
		return createErr
	})
	if err != nil {
		http.Error(w, "Failed to create highlight: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.db.UpdateAPIKeyLastUsed(apiKey.ID)

	selectorJSON, _ := json.Marshal(req.Selector)
	selectorStr := string(selectorJSON)
	colorPtr := &color

	highlight := &db.Highlight{
		URI:          result.URI,
		AuthorDID:    apiKey.OwnerDID,
		TargetSource: req.URL,
		TargetHash:   urlHash,
		SelectorJSON: &selectorStr,
		Color:        colorPtr,
		CreatedAt:    time.Now(),
		IndexedAt:    time.Now(),
		CID:          &result.CID,
	}
	if err := h.db.CreateHighlight(highlight); err != nil {
		fmt.Printf("Warning: failed to index highlight in local DB: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"uri":     result.URI,
		"cid":     result.CID,
		"message": "Highlight created successfully",
	})
}

func (h *APIKeyHandler) authenticateAPIKey(r *http.Request) (*db.APIKey, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, fmt.Errorf("invalid Authorization format, expected 'Bearer <key>'")
	}

	rawKey := strings.TrimPrefix(auth, "Bearer ")
	keyHash := hashAPIKey(rawKey)

	apiKey, err := h.db.GetAPIKeyByHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("invalid API key")
	}

	return apiKey, nil
}

func (h *APIKeyHandler) getSessionByDID(did string) (*SessionData, error) {
	rows, err := h.db.Query(h.db.Rebind(`
		SELECT id, did, handle, access_token, refresh_token, COALESCE(dpop_key, '')
		FROM sessions
		WHERE did = ? AND expires_at > ?
		ORDER BY created_at DESC
		LIMIT 1
	`), did, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no active session")
	}

	var sessionID, sessDID, handle, accessToken, refreshToken, dpopKeyStr string
	if err := rows.Scan(&sessionID, &sessDID, &handle, &accessToken, &refreshToken, &dpopKeyStr); err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(dpopKeyStr))
	if block == nil {
		return nil, fmt.Errorf("invalid session DPoP key")
	}
	dpopKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid session DPoP key: %w", err)
	}

	pds, err := xrpc.ResolveDIDToPDS(sessDID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve PDS: %w", err)
	}
	if pds == "" {
		return nil, fmt.Errorf("PDS not found for DID: %s", sessDID)
	}

	return &SessionData{
		ID:           sessionID,
		DID:          sessDID,
		Handle:       handle,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DPoPKey:      dpopKey,
		PDS:          pds,
	}, nil
}

func generateAPIKey() string {
	b := make([]byte, 32)
	rand.Read(b)
	return "mk_" + hex.EncodeToString(b)
}

func generateKeyID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashAPIKey(key string) string {
	h := sha256.New()
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}
