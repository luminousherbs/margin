package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"margin.at/internal/db"
	"margin.at/internal/xrpc"
)

type AnnotationService struct {
	db        *db.DB
	refresher *TokenRefresher
}

func NewAnnotationService(database *db.DB, refresher *TokenRefresher) *AnnotationService {
	return &AnnotationService{db: database, refresher: refresher}
}

type CreateAnnotationRequest struct {
	URL      string          `json:"url"`
	Text     string          `json:"text"`
	Selector json.RawMessage `json:"selector,omitempty"`
	Title    string          `json:"title,omitempty"`
	Tags     []string        `json:"tags,omitempty"`
	Labels   []string        `json:"labels,omitempty"`
}

type CreateAnnotationResponse struct {
	URI string `json:"uri"`
	CID string `json:"cid"`
}

func (s *AnnotationService) CreateAnnotation(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if req.Text == "" && req.Selector == nil && len(req.Tags) == 0 {
		http.Error(w, "Must provide text, selector, or tags", http.StatusBadRequest)
		return
	}

	if len(req.Text) > 3000 {
		http.Error(w, "Text too long (max 3000 chars)", http.StatusBadRequest)
		return
	}

	urlHash := db.HashURL(req.URL)

	motivation := "commenting"
	if req.Selector != nil && req.Text == "" {
		motivation = "highlighting"
	} else if len(req.Tags) > 0 {
		motivation = "tagging"
	}

	var facets []xrpc.Facet
	var mentionedDIDs []string

	mentionRegex := regexp.MustCompile(`(^|\s|@)@([a-zA-Z0-9.-]+)(\b)`)
	matches := mentionRegex.FindAllStringSubmatchIndex(req.Text, -1)

	for _, m := range matches {
		handle := req.Text[m[4]:m[5]]

		if !strings.Contains(handle, ".") {
			continue
		}

		var did string
		err := s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, _ string) error {
			var resolveErr error
			did, resolveErr = client.ResolveHandle(r.Context(), handle)
			return resolveErr
		})

		if err == nil && did != "" {
			start := m[2]
			end := m[5]

			facets = append(facets, xrpc.Facet{
				Index: xrpc.FacetIndex{
					ByteStart: start,
					ByteEnd:   end,
				},
				Features: []xrpc.FacetFeature{
					{
						Type: "app.bsky.richtext.facet#mention",
						Did:  did,
					},
				},
			})
			mentionedDIDs = append(mentionedDIDs, did)
		}
	}

	urlRegex := regexp.MustCompile(`(https?://[^\s]+)`)
	urlMatches := urlRegex.FindAllStringIndex(req.Text, -1)

	for _, m := range urlMatches {
		facets = append(facets, xrpc.Facet{
			Index: xrpc.FacetIndex{
				ByteStart: m[0],
				ByteEnd:   m[1],
			},
			Features: []xrpc.FacetFeature{
				{
					Type: "app.bsky.richtext.facet#link",
					Uri:  req.Text[m[0]:m[1]],
				},
			},
		})
	}

	record := xrpc.NewAnnotationRecordWithMotivation(req.URL, urlHash, req.Text, req.Selector, req.Title, motivation)
	if len(req.Tags) > 0 {
		record.Tags = req.Tags
	}
	if len(facets) > 0 {
		record.Facets = facets
	}

	validSelfLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
	var validLabels []string
	for _, l := range req.Labels {
		if validSelfLabels[l] {
			validLabels = append(validLabels, l)
		}
	}
	record.Labels = xrpc.NewSelfLabels(validLabels)

	var result *xrpc.CreateRecordOutput

	if existing, err := s.checkDuplicateAnnotation(session.DID, req.URL, req.Text); err == nil && existing != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreateAnnotationResponse{
			URI: existing.URI,
			CID: *existing.CID,
		})
		return
	}

	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionAnnotation, record)
		return createErr
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to create annotation: ", http.StatusInternalServerError)
		return
	}

	for _, mentionedDID := range mentionedDIDs {
		if mentionedDID != session.DID {
			s.db.CreateNotification(&db.Notification{
				RecipientDID: mentionedDID,
				ActorDID:     session.DID,
				Type:         "mention",
				SubjectURI:   result.URI,
				CreatedAt:    time.Now(),
			})
		}
	}

	bodyValue := req.Text
	var bodyValuePtr, targetTitlePtr, selectorJSONPtr *string
	if bodyValue != "" {
		bodyValuePtr = &bodyValue
	}
	if req.Title != "" {
		targetTitlePtr = &req.Title
	}
	if req.Selector != nil {
		selectorBytes, _ := json.Marshal(req.Selector)
		selectorStr := string(selectorBytes)
		selectorJSONPtr = &selectorStr
	}

	var tagsJSONPtr *string
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		tagsJSONPtr = &tagsStr
	}

	cid := result.CID
	did := session.DID
	annotation := &db.Annotation{
		URI:          result.URI,
		CID:          &cid,
		AuthorDID:    did,
		Motivation:   motivation,
		BodyValue:    bodyValuePtr,
		TargetSource: req.URL,
		TargetHash:   urlHash,
		TargetTitle:  targetTitlePtr,
		SelectorJSON: selectorJSONPtr,
		TagsJSON:     tagsJSONPtr,
		CreatedAt:    time.Now(),
		IndexedAt:    time.Now(),
	}

	if err := s.db.CreateAnnotation(annotation); err != nil {
		log.Printf("Warning: failed to index annotation in local DB: %v", err)
	}

	for _, label := range validLabels {
		if err := s.db.CreateContentLabel(session.DID, result.URI, label, session.DID); err != nil {
			log.Printf("Warning: failed to create self-label %s: %v", label, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreateAnnotationResponse{
		URI: result.URI,
		CID: result.CID,
	})
}

func (s *AnnotationService) DeleteAnnotation(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	rkey := r.URL.Query().Get("rkey")
	collectionType := r.URL.Query().Get("type")

	if rkey == "" {
		http.Error(w, "rkey required", http.StatusBadRequest)
		return
	}

	did := session.DID

	collection := xrpc.CollectionAnnotation
	if collectionType == "reply" {
		collection = xrpc.CollectionReply
	} else {
		candidateCollections := []string{xrpc.CollectionAnnotation, "network.cosmik.card"}
		for _, col := range candidateCollections {
			uri := "at://" + did + "/" + col + "/" + rkey
			if _, dbErr := s.db.GetAnnotationByURI(uri); dbErr == nil {
				collection = col
				break
			}
		}
	}

	pdsErr := s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		return client.DeleteRecord(r.Context(), did, collection, rkey)
	})
	if pdsErr != nil {
		log.Printf("PDS delete failed (will still clean local DB): %v", pdsErr)
	}

	// Always clean up local DB regardless of PDS result
	if collectionType == "reply" {
		uri := "at://" + did + "/" + xrpc.CollectionReply + "/" + rkey
		s.db.DeleteReply(uri)
	} else {
		uri := "at://" + did + "/" + collection + "/" + rkey
		s.db.DeleteAnnotation(uri)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

type UpdateAnnotationRequest struct {
	Text   string   `json:"text"`
	Tags   []string `json:"tags"`
	Labels []string `json:"labels,omitempty"`
}

func (s *AnnotationService) UpdateAnnotation(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	annotation, err := s.db.GetAnnotationByURI(uri)
	if err != nil || annotation == nil {
		http.Error(w, "Annotation not found", http.StatusNotFound)
		return
	}

	if annotation.AuthorDID != session.DID {
		http.Error(w, "Not authorized to edit this annotation", http.StatusForbidden)
		return
	}

	var req UpdateAnnotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	parts := parseATURI(uri)
	if len(parts) < 3 {
		http.Error(w, "Invalid URI format", http.StatusBadRequest)
		return
	}
	rkey := parts[2]

	tagsJSON := ""
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsJSON = string(tagsBytes)
	}

	if annotation.BodyValue != nil {
		previousContent := *annotation.BodyValue
		log.Printf("[DEBUG] Saving edit history for %s. Previous content: %s", uri, previousContent)
		if err := s.db.SaveEditHistory(uri, "annotation", previousContent, annotation.CID); err != nil {
			log.Printf("Failed to save edit history for %s: %v", uri, err)
		} else {
			log.Printf("[DEBUG] Successfully saved edit history for %s", uri)
		}
	} else {
		log.Printf("[DEBUG] Annotation BodyValue is nil for %s", uri)
	}

	var result *xrpc.PutRecordOutput
	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		existing, getErr := client.GetRecord(r.Context(), did, xrpc.CollectionAnnotation, rkey)
		if getErr != nil {
			return fmt.Errorf("failed to fetch existing record: %w", getErr)
		}

		var record xrpc.AnnotationRecord
		if err := json.Unmarshal(existing.Value, &record); err != nil {
			return fmt.Errorf("failed to parse existing record: %w", err)
		}

		record.Body = &xrpc.AnnotationBody{
			Value:  req.Text,
			Format: "text/plain",
		}
		if len(req.Tags) > 0 {
			record.Tags = req.Tags
		} else {
			record.Tags = nil
		}

		updateValidLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
		var updateLabels []string
		for _, l := range req.Labels {
			if updateValidLabels[l] {
				updateLabels = append(updateLabels, l)
			}
		}
		record.Labels = xrpc.NewSelfLabels(updateLabels)

		if err := record.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		var updateErr error
		result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionAnnotation, rkey, record)
		if updateErr != nil {
			log.Printf("UpdateAnnotation failed: %v. Retrying with delete-then-create workaround.", updateErr)
			_ = client.DeleteRecord(r.Context(), did, xrpc.CollectionAnnotation, rkey)
			result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionAnnotation, rkey, record)
		}
		return updateErr
	})

	if err != nil {
		log.Printf("[UpdateAnnotation] Failed: %v", err)
		HandleAPIError(w, r, err, "Failed to update record: ", http.StatusInternalServerError)
		return
	}

	s.db.UpdateAnnotation(uri, req.Text, tagsJSON, result.CID)

	validSelfLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
	var validLabels []string
	for _, l := range req.Labels {
		if validSelfLabels[l] {
			validLabels = append(validLabels, l)
		}
	}
	if err := s.db.SyncSelfLabels(session.DID, uri, validLabels); err != nil {
		log.Printf("Warning: failed to sync self-labels: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"uri":     result.URI,
		"cid":     result.CID,
	})
}

func parseATURI(uri string) []string {

	if len(uri) < 5 || uri[:5] != "at://" {
		return nil
	}
	return strings.Split(uri[5:], "/")
}

type CreateLikeRequest struct {
	SubjectURI string `json:"subjectUri"`
	SubjectCID string `json:"subjectCid"`
}

func (s *AnnotationService) LikeAnnotation(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateLikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SubjectURI == "" || req.SubjectCID == "" {
		http.Error(w, "subjectUri and subjectCid are required", http.StatusBadRequest)
		return
	}

	existingLike, _ := s.db.GetLikeByUserAndSubject(session.DID, req.SubjectURI)
	if existingLike != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"uri": existingLike.URI, "existing": "true"})
		return
	}

	record := xrpc.NewLikeRecord(req.SubjectURI, req.SubjectCID)

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput
	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionLike, record)
		return createErr
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to create like: ", http.StatusInternalServerError)
		return
	}

	did := session.DID
	like := &db.Like{
		URI:        result.URI,
		AuthorDID:  did,
		SubjectURI: req.SubjectURI,
		CreatedAt:  time.Now(),
		IndexedAt:  time.Now(),
	}
	s.db.CreateLike(like)

	if authorDID, err := s.db.GetAuthorByURI(req.SubjectURI); err == nil && authorDID != did {
		s.db.CreateNotification(&db.Notification{
			RecipientDID: authorDID,
			ActorDID:     did,
			Type:         "like",
			SubjectURI:   req.SubjectURI,
			CreatedAt:    time.Now(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"uri": result.URI})
}

func (s *AnnotationService) UnlikeAnnotation(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	subjectURI := r.URL.Query().Get("uri")
	if subjectURI == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	userLike, err := s.db.GetLikeByUserAndSubject(session.DID, subjectURI)
	if err != nil {
		http.Error(w, "Like not found", http.StatusNotFound)
		return
	}

	parts := strings.Split(userLike.URI, "/")
	rkey := parts[len(parts)-1]

	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		return client.DeleteRecord(r.Context(), did, xrpc.CollectionLike, rkey)
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to delete like: ", http.StatusInternalServerError)
		return
	}

	s.db.DeleteLike(userLike.URI)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

type CreateReplyRequest struct {
	ParentURI string `json:"parentUri"`
	ParentCID string `json:"parentCid"`
	RootURI   string `json:"rootUri"`
	RootCID   string `json:"rootCid"`
	Text      string `json:"text"`
}

func (s *AnnotationService) CreateReply(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ParentURI == "" || req.ParentCID == "" {
		http.Error(w, "parentUri and parentCid are required", http.StatusBadRequest)
		return
	}
	if req.RootURI == "" || req.RootCID == "" {
		http.Error(w, "rootUri and rootCid are required", http.StatusBadRequest)
		return
	}
	if req.Text == "" {
		http.Error(w, "text is required", http.StatusBadRequest)
		return
	}

	record := xrpc.NewReplyRecord(req.ParentURI, req.ParentCID, req.RootURI, req.RootCID, req.Text)

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput
	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionReply, record)
		return createErr
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to create reply: ", http.StatusInternalServerError)
		return
	}

	reply := &db.Reply{
		URI:       result.URI,
		AuthorDID: session.DID,
		ParentURI: req.ParentURI,
		RootURI:   req.RootURI,
		Text:      req.Text,
		CreatedAt: time.Now(),
		IndexedAt: time.Now(),
		CID:       &result.CID,
	}
	s.db.CreateReply(reply)

	if authorDID, err := s.db.GetAuthorByURI(req.ParentURI); err == nil && authorDID != session.DID {
		s.db.CreateNotification(&db.Notification{
			RecipientDID: authorDID,
			ActorDID:     session.DID,
			Type:         "reply",
			SubjectURI:   result.URI,
			CreatedAt:    time.Now(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"uri": result.URI})
}

func (s *AnnotationService) DeleteReply(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	reply, err := s.db.GetReplyByURI(uri)
	if err != nil || reply == nil {
		http.Error(w, "reply not found", http.StatusNotFound)
		return
	}

	if reply.AuthorDID != session.DID {
		http.Error(w, "not authorized to delete this reply", http.StatusForbidden)
		return
	}

	parts := strings.Split(uri, "/")
	if len(parts) >= 2 {
		rkey := parts[len(parts)-1]
		_ = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
			return client.DeleteRecord(r.Context(), did, "at.margin.reply", rkey)
		})
	}

	s.db.DeleteReply(uri)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

type CreateHighlightRequest struct {
	URL      string          `json:"url"`
	Title    string          `json:"title,omitempty"`
	Selector json.RawMessage `json:"selector"`
	Color    string          `json:"color,omitempty"`
	Tags     []string        `json:"tags,omitempty"`
	Labels   []string        `json:"labels,omitempty"`
}

func (s *AnnotationService) CreateHighlight(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateHighlightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" || req.Selector == nil {
		http.Error(w, "URL and selector are required", http.StatusBadRequest)
		return
	}

	urlHash := db.HashURL(req.URL)
	record := xrpc.NewHighlightRecord(req.URL, urlHash, req.Selector, req.Color, req.Tags)

	validSelfLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
	var validLabels []string
	for _, l := range req.Labels {
		if validSelfLabels[l] {
			validLabels = append(validLabels, l)
		}
	}
	record.Labels = xrpc.NewSelfLabels(validLabels)

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput

	if existing, err := s.checkDuplicateHighlight(session.DID, req.URL, req.Selector); err == nil && existing != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"uri": existing.URI, "cid": *existing.CID})
		return
	}

	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionHighlight, record)
		return createErr
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to create highlight: ", http.StatusInternalServerError)
		return
	}

	var selectorJSONPtr *string
	if len(record.Target.Selector) > 0 {
		selectorStr := string(record.Target.Selector)
		selectorJSONPtr = &selectorStr
	}

	var titlePtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}

	var colorPtr *string
	if req.Color != "" {
		colorPtr = &req.Color
	}

	var tagsJSONPtr *string
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		tagsJSONPtr = &tagsStr
	}

	cid := result.CID
	highlight := &db.Highlight{
		URI:          result.URI,
		AuthorDID:    session.DID,
		TargetSource: req.URL,
		TargetHash:   urlHash,
		TargetTitle:  titlePtr,
		SelectorJSON: selectorJSONPtr,
		Color:        colorPtr,
		TagsJSON:     tagsJSONPtr,
		CreatedAt:    time.Now(),
		IndexedAt:    time.Now(),
		CID:          &cid,
	}
	if err := s.db.CreateHighlight(highlight); err != nil {
		http.Error(w, "Failed to index highlight", http.StatusInternalServerError)
		return
	}

	for _, label := range validLabels {
		if err := s.db.CreateContentLabel(session.DID, result.URI, label, session.DID); err != nil {
			log.Printf("Warning: failed to create self-label %s: %v", label, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"uri": result.URI, "cid": result.CID})
}

type CreateBookmarkRequest struct {
	URL         string   `json:"url"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func (s *AnnotationService) CreateBookmark(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req CreateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	urlHash := db.HashURL(req.URL)
	record := xrpc.NewBookmarkRecord(req.URL, urlHash, req.Title, req.Description)
	if len(req.Tags) > 0 {
		record.Tags = req.Tags
	}

	if err := record.Validate(); err != nil {
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var result *xrpc.CreateRecordOutput

	if existing, err := s.checkDuplicateBookmark(session.DID, req.URL); err == nil && existing != nil {
		http.Error(w, "Bookmark already exists", http.StatusConflict)
		return
	}

	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		var createErr error
		result, createErr = client.CreateRecord(r.Context(), did, xrpc.CollectionBookmark, record)
		return createErr
	})
	if err != nil {
		HandleAPIError(w, r, err, "Failed to create bookmark: ", http.StatusInternalServerError)
		return
	}

	var titlePtr *string
	if req.Title != "" {
		titlePtr = &req.Title
	}
	var descPtr *string
	if req.Description != "" {
		descPtr = &req.Description
	}

	var tagsJSONPtr *string
	if len(req.Tags) > 0 {
		tagsBytes, _ := json.Marshal(req.Tags)
		tagsStr := string(tagsBytes)
		tagsJSONPtr = &tagsStr
	}

	cid := result.CID
	bookmark := &db.Bookmark{
		URI:         result.URI,
		AuthorDID:   session.DID,
		Source:      req.URL,
		SourceHash:  urlHash,
		Title:       titlePtr,
		Description: descPtr,
		TagsJSON:    tagsJSONPtr,
		CreatedAt:   time.Now(),
		IndexedAt:   time.Now(),
		CID:         &cid,
	}
	s.db.CreateBookmark(bookmark)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"uri": result.URI, "cid": result.CID})
}

func (s *AnnotationService) DeleteHighlight(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	rkey := r.URL.Query().Get("rkey")
	if rkey == "" {
		http.Error(w, "rkey required", http.StatusBadRequest)
		return
	}

	pdsErr := s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		return client.DeleteRecord(r.Context(), did, xrpc.CollectionHighlight, rkey)
	})
	if pdsErr != nil {
		log.Printf("PDS delete highlight failed (will still clean local DB): %v", pdsErr)
	}

	uri := "at://" + session.DID + "/" + xrpc.CollectionHighlight + "/" + rkey
	s.db.DeleteHighlight(uri)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (s *AnnotationService) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	rkey := r.URL.Query().Get("rkey")
	if rkey == "" {
		http.Error(w, "rkey required", http.StatusBadRequest)
		return
	}

	pdsErr := s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		return client.DeleteRecord(r.Context(), did, xrpc.CollectionBookmark, rkey)
	})
	if pdsErr != nil {
		log.Printf("PDS delete bookmark failed (will still clean local DB): %v", pdsErr)
	}

	uri := "at://" + session.DID + "/" + xrpc.CollectionBookmark + "/" + rkey
	s.db.DeleteBookmark(uri)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

type UpdateHighlightRequest struct {
	Color  string   `json:"color"`
	Tags   []string `json:"tags,omitempty"`
	Labels []string `json:"labels,omitempty"`
}

func (s *AnnotationService) UpdateHighlight(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if len(uri) < 5 || !strings.HasPrefix(uri[5:], session.DID) {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	var req UpdateHighlightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	parts := parseATURI(uri)
	if len(parts) < 3 {
		http.Error(w, "Invalid URI", http.StatusBadRequest)
		return
	}
	rkey := parts[2]

	var result *xrpc.PutRecordOutput
	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		existing, getErr := client.GetRecord(r.Context(), did, xrpc.CollectionHighlight, rkey)
		if getErr != nil {
			return fmt.Errorf("failed to fetch record: %w", getErr)
		}

		var record xrpc.HighlightRecord
		json.Unmarshal(existing.Value, &record)

		if req.Color != "" {
			record.Color = req.Color
		}
		if req.Tags != nil {
			record.Tags = req.Tags
		}

		updateValidLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
		var updateLabels []string
		for _, l := range req.Labels {
			if updateValidLabels[l] {
				updateLabels = append(updateLabels, l)
			}
		}
		record.Labels = xrpc.NewSelfLabels(updateLabels)

		if err := record.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		var updateErr error
		result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionHighlight, rkey, record)
		if updateErr != nil {
			log.Printf("UpdateHighlight failed: %v. Retrying with delete-then-create workaround.", updateErr)
			_ = client.DeleteRecord(r.Context(), did, xrpc.CollectionHighlight, rkey)
			result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionHighlight, rkey, record)
		}
		return updateErr
	})

	if err != nil {
		HandleAPIError(w, r, err, "Failed to update: ", http.StatusInternalServerError)
		return
	}

	tagsJSON := ""
	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		tagsJSON = string(b)
	}
	s.db.UpdateHighlight(uri, req.Color, tagsJSON, result.CID)

	validSelfLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
	var validLabels []string
	for _, l := range req.Labels {
		if validSelfLabels[l] {
			validLabels = append(validLabels, l)
		}
	}
	if err := s.db.SyncSelfLabels(session.DID, uri, validLabels); err != nil {
		log.Printf("Warning: failed to sync self-labels: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "uri": result.URI, "cid": result.CID})
}

type UpdateBookmarkRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags,omitempty"`
	Labels      []string `json:"labels,omitempty"`
}

func (s *AnnotationService) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	session, err := s.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if len(uri) < 5 || !strings.HasPrefix(uri[5:], session.DID) {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}

	var req UpdateBookmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	parts := parseATURI(uri)
	if len(parts) < 3 {
		http.Error(w, "Invalid URI", http.StatusBadRequest)
		return
	}
	rkey := parts[2]

	var result *xrpc.PutRecordOutput
	err = s.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, did string) error {
		existing, getErr := client.GetRecord(r.Context(), did, xrpc.CollectionBookmark, rkey)
		if getErr != nil {
			return fmt.Errorf("failed to fetch record: %w", getErr)
		}

		var record xrpc.BookmarkRecord
		json.Unmarshal(existing.Value, &record)

		if req.Title != "" {
			record.Title = req.Title
		}
		if req.Description != "" {
			record.Description = req.Description
		}
		if req.Tags != nil {
			record.Tags = req.Tags
		}

		updateValidLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
		var updateLabels []string
		for _, l := range req.Labels {
			if updateValidLabels[l] {
				updateLabels = append(updateLabels, l)
			}
		}
		record.Labels = xrpc.NewSelfLabels(updateLabels)

		if err := record.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		var updateErr error
		result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionBookmark, rkey, record)
		if updateErr != nil {
			log.Printf("UpdateBookmark failed: %v. Retrying with delete-then-create workaround.", updateErr)
			_ = client.DeleteRecord(r.Context(), did, xrpc.CollectionBookmark, rkey)
			result, updateErr = client.PutRecord(r.Context(), did, xrpc.CollectionBookmark, rkey, record)
		}
		return updateErr
	})

	if err != nil {
		HandleAPIError(w, r, err, "Failed to update: ", http.StatusInternalServerError)
		return
	}

	tagsJSON := ""
	if req.Tags != nil {
		b, _ := json.Marshal(req.Tags)
		tagsJSON = string(b)
	}
	s.db.UpdateBookmark(uri, req.Title, req.Description, tagsJSON, result.CID)

	validSelfLabels := map[string]bool{"sexual": true, "nudity": true, "violence": true, "gore": true, "spam": true, "misleading": true}
	var validLabels []string
	for _, l := range req.Labels {
		if validSelfLabels[l] {
			validLabels = append(validLabels, l)
		}
	}
	if err := s.db.SyncSelfLabels(session.DID, uri, validLabels); err != nil {
		log.Printf("Warning: failed to sync self-labels: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "uri": result.URI, "cid": result.CID})
}
