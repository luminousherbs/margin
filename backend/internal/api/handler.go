package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"margin.at/internal/db"
	internal_sync "margin.at/internal/sync"
	"margin.at/internal/xrpc"
)

type Handler struct {
	db                *db.DB
	annotationService *AnnotationService
	refresher         *TokenRefresher
	apiKeys           *APIKeyHandler
	syncService       *internal_sync.Service
	moderation        *ModerationHandler
}

func NewHandler(database *db.DB, annotationService *AnnotationService, refresher *TokenRefresher, syncService *internal_sync.Service) *Handler {
	return &Handler{
		db:                database,
		annotationService: annotationService,
		refresher:         refresher,
		apiKeys:           NewAPIKeyHandler(database, refresher),
		syncService:       syncService,
		moderation:        NewModerationHandler(database, refresher),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.Health)

	r.Route("/api", func(r chi.Router) {
		r.Get("/annotations", h.GetAnnotations)
		r.Get("/annotations/feed", h.GetFeed)
		r.Get("/annotation", h.GetAnnotation)
		r.Get("/annotations/history", h.GetEditHistory)
		r.Put("/annotations", h.annotationService.UpdateAnnotation)

		r.Get("/highlights", h.GetHighlights)
		r.Put("/highlights", h.annotationService.UpdateHighlight)

		r.Get("/bookmarks", h.GetBookmarks)
		r.Post("/bookmarks", h.annotationService.CreateBookmark)
		r.Put("/bookmarks", h.annotationService.UpdateBookmark)

		collectionService := NewCollectionService(h.db, h.refresher)
		r.Post("/collections", collectionService.CreateCollection)
		r.Get("/collections", collectionService.GetCollections)
		r.Put("/collections", collectionService.UpdateCollection)
		r.Delete("/collections", collectionService.DeleteCollection)
		r.Post("/collections/{collection}/items", collectionService.AddCollectionItem)
		r.Get("/collections/{collection}/items", collectionService.GetCollectionItems)
		r.Delete("/collections/items", collectionService.RemoveCollectionItem)
		r.Get("/collections/containing", collectionService.GetAnnotationCollections)
		r.Get("/collection", collectionService.GetCollection)
		r.Post("/sync", h.SyncAll)

		r.Get("/targets", h.GetByTarget)
		r.Get("/discover", h.DiscoverForURL)

		r.Get("/users/{did}/annotations", h.GetUserAnnotations)
		r.Get("/users/{did}/highlights", h.GetUserHighlights)
		r.Get("/users/{did}/bookmarks", h.GetUserBookmarks)
		r.Get("/users/{did}/targets", h.GetUserTargetItems)
		r.Get("/users/{did}/tags", h.HandleGetUserTags)

		r.Get("/trending-tags", h.HandleGetTrendingTags)

		r.Get("/replies", h.GetReplies)
		r.Get("/likes", h.GetLikeCount)
		r.Get("/url-metadata", h.GetURLMetadata)
		r.Get("/notifications", h.GetNotifications)
		r.Get("/notifications/count", h.GetUnreadNotificationCount)
		r.Post("/notifications/read", h.MarkNotificationsRead)
		r.Get("/avatar/{did}", h.HandleAvatarProxy)

		r.Post("/keys", h.apiKeys.CreateKey)
		r.Get("/keys", h.apiKeys.ListKeys)
		r.Delete("/keys/{id}", h.apiKeys.DeleteKey)

		r.Post("/quick/bookmark", h.apiKeys.QuickBookmark)
		r.Post("/quick/save", h.apiKeys.QuickSave)

		r.Get("/preferences", h.GetPreferences)
		r.Put("/preferences", h.UpdatePreferences)

		r.Post("/moderation/block", h.moderation.BlockUser)
		r.Delete("/moderation/block", h.moderation.UnblockUser)
		r.Get("/moderation/blocks", h.moderation.GetBlocks)
		r.Post("/moderation/mute", h.moderation.MuteUser)
		r.Delete("/moderation/mute", h.moderation.UnmuteUser)
		r.Get("/moderation/mutes", h.moderation.GetMutes)
		r.Get("/moderation/relationship", h.moderation.GetRelationship)
		r.Post("/moderation/report", h.moderation.CreateReport)
		r.Get("/moderation/admin/check", h.moderation.AdminCheckAccess)
		r.Get("/moderation/admin/reports", h.moderation.AdminGetReports)
		r.Get("/moderation/admin/report", h.moderation.AdminGetReport)
		r.Post("/moderation/admin/action", h.moderation.AdminTakeAction)
		r.Post("/moderation/admin/label", h.moderation.AdminCreateLabel)
		r.Delete("/moderation/admin/label", h.moderation.AdminDeleteLabel)
		r.Get("/moderation/admin/labels", h.moderation.AdminGetLabels)
		r.Get("/moderation/labeler", h.moderation.GetLabelerInfo)
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "1.0"})
}

func (h *Handler) GetAnnotations(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		source = r.URL.Query().Get("url")
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)
	motivation := r.URL.Query().Get("motivation")
	tag := r.URL.Query().Get("tag")

	var annotations []db.Annotation
	var err error

	if source != "" {
		urlHash := db.HashURL(source)
		annotations, err = h.db.GetAnnotationsByTargetHash(urlHash, limit, offset)
	} else if motivation != "" {
		annotations, err = h.db.GetAnnotationsByMotivation(motivation, limit, offset)
	} else if tag != "" {
		annotations, err = h.db.GetAnnotationsByTag(tag, limit, offset)
	} else {
		annotations, err = h.db.GetRecentAnnotations(limit, offset)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateAnnotations(h.db, annotations, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "AnnotationCollection",
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetFeed(w http.ResponseWriter, r *http.Request) {
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)
	tag := r.URL.Query().Get("tag")
	creator := r.URL.Query().Get("creator")
	feedType := r.URL.Query().Get("type")

	viewerDID := h.getViewerDID(r)

	var annotations []db.Annotation
	var highlights []db.Highlight
	var bookmarks []db.Bookmark
	var collectionItems []db.CollectionItem
	var err error

	motivation := r.URL.Query().Get("motivation")

	fetchLimit := limit + offset

	if tag != "" {
		if creator != "" {
			if motivation == "" || motivation == "commenting" {
				switch feedType {
				case "margin":
					annotations, _ = h.db.GetMarginAnnotationsByTagAndAuthor(tag, creator, fetchLimit, 0)
				case "semble":
					annotations, _ = h.db.GetSembleAnnotationsByTagAndAuthor(tag, creator, fetchLimit, 0)
				default:
					annotations, _ = h.db.GetAnnotationsByTagAndAuthor(tag, creator, fetchLimit, 0)
				}
			}
			if motivation == "" || motivation == "highlighting" {
				switch feedType {
				case "margin":
					highlights, _ = h.db.GetMarginHighlightsByTagAndAuthor(tag, creator, fetchLimit, 0)
				case "semble":
					highlights, _ = h.db.GetSembleHighlightsByTagAndAuthor(tag, creator, fetchLimit, 0)
				default:
					highlights, _ = h.db.GetHighlightsByTagAndAuthor(tag, creator, fetchLimit, 0)
				}
			}
			if motivation == "" || motivation == "bookmarking" {
				switch feedType {
				case "margin":
					bookmarks, _ = h.db.GetMarginBookmarksByTagAndAuthor(tag, creator, fetchLimit, 0)
				case "semble":
					bookmarks, _ = h.db.GetSembleBookmarksByTagAndAuthor(tag, creator, fetchLimit, 0)
				default:
					bookmarks, _ = h.db.GetBookmarksByTagAndAuthor(tag, creator, fetchLimit, 0)
				}
			}
			collectionItems = []db.CollectionItem{}
		} else {
			if motivation == "" || motivation == "commenting" {
				switch feedType {
				case "margin":
					annotations, _ = h.db.GetMarginAnnotationsByTag(tag, fetchLimit, 0)
				case "semble":
					annotations, _ = h.db.GetSembleAnnotationsByTag(tag, fetchLimit, 0)
				default:
					annotations, _ = h.db.GetAnnotationsByTag(tag, fetchLimit, 0)
				}
			}
			if motivation == "" || motivation == "highlighting" {
				switch feedType {
				case "margin":
					highlights, _ = h.db.GetMarginHighlightsByTag(tag, fetchLimit, 0)
				case "semble":
					highlights, _ = h.db.GetSembleHighlightsByTag(tag, fetchLimit, 0)
				default:
					highlights, _ = h.db.GetHighlightsByTag(tag, fetchLimit, 0)
				}
			}
			if motivation == "" || motivation == "bookmarking" {
				switch feedType {
				case "margin":
					bookmarks, _ = h.db.GetMarginBookmarksByTag(tag, fetchLimit, 0)
				case "semble":
					bookmarks, _ = h.db.GetSembleBookmarksByTag(tag, fetchLimit, 0)
				default:
					bookmarks, _ = h.db.GetBookmarksByTag(tag, fetchLimit, 0)
				}
			}
			collectionItems = []db.CollectionItem{}
		}
	} else if creator != "" {
		if motivation == "" || motivation == "commenting" {
			switch feedType {
			case "margin":
				annotations, _ = h.db.GetMarginAnnotationsByAuthor(creator, fetchLimit, 0)
			case "semble":
				annotations, _ = h.db.GetSembleAnnotationsByAuthor(creator, fetchLimit, 0)
			default:
				annotations, _ = h.db.GetAnnotationsByAuthor(creator, fetchLimit, 0)
			}
		}
		if motivation == "" || motivation == "highlighting" {
			switch feedType {
			case "margin":
				highlights, _ = h.db.GetMarginHighlightsByAuthor(creator, fetchLimit, 0)
			case "semble":
				highlights, _ = h.db.GetSembleHighlightsByAuthor(creator, fetchLimit, 0)
			default:
				highlights, _ = h.db.GetHighlightsByAuthor(creator, fetchLimit, 0)
			}
		}
		if motivation == "" || motivation == "bookmarking" {
			switch feedType {
			case "margin":
				bookmarks, _ = h.db.GetMarginBookmarksByAuthor(creator, fetchLimit, 0)
			case "semble":
				bookmarks, _ = h.db.GetSembleBookmarksByAuthor(creator, fetchLimit, 0)
			default:
				bookmarks, _ = h.db.GetBookmarksByAuthor(creator, fetchLimit, 0)
			}
		}
		collectionItems = []db.CollectionItem{}
	} else {
		if motivation == "" || motivation == "commenting" {
			switch feedType {
			case "margin":
				annotations, _ = h.db.GetMarginAnnotations(fetchLimit, 0)
			case "semble":
				annotations, _ = h.db.GetSembleAnnotations(fetchLimit, 0)
			case "popular":
				annotations, _ = h.db.GetPopularAnnotations(fetchLimit, 0)
			case "shelved":
				annotations, _ = h.db.GetShelvedAnnotations(fetchLimit, 0)
			default:
				annotations, _ = h.db.GetRecentAnnotations(fetchLimit, 0)
			}
		}
		if motivation == "" || motivation == "highlighting" {
			switch feedType {
			case "margin":
				highlights, _ = h.db.GetMarginHighlights(fetchLimit, 0)
			case "semble":
				highlights, _ = h.db.GetSembleHighlights(fetchLimit, 0)
			case "popular":
				highlights, _ = h.db.GetPopularHighlights(fetchLimit, 0)
			case "shelved":
				highlights, _ = h.db.GetShelvedHighlights(fetchLimit, 0)
			default:
				highlights, _ = h.db.GetRecentHighlights(fetchLimit, 0)
			}
		}
		if motivation == "" || motivation == "bookmarking" {
			switch feedType {
			case "margin":
				bookmarks, _ = h.db.GetMarginBookmarks(fetchLimit, 0)
			case "semble":
				bookmarks, _ = h.db.GetSembleBookmarks(fetchLimit, 0)
			case "popular":
				bookmarks, _ = h.db.GetPopularBookmarks(fetchLimit, 0)
			case "shelved":
				bookmarks, _ = h.db.GetShelvedBookmarks(fetchLimit, 0)
			default:
				bookmarks, _ = h.db.GetRecentBookmarks(fetchLimit, 0)
			}
		}
		if motivation == "" {
			switch feedType {
			case "popular":
				collectionItems, err = h.db.GetPopularCollectionItems(fetchLimit, 0)
			case "shelved":
				collectionItems, err = h.db.GetShelvedCollectionItems(fetchLimit, 0)
			default:
				collectionItems, err = h.db.GetRecentCollectionItems(fetchLimit, 0)
			}
			if err != nil {
				log.Printf("Error fetching collection items: %v\n", err)
			}
		}
	}

	authAnnos, _ := hydrateAnnotations(h.db, annotations, viewerDID)
	authHighs, _ := hydrateHighlights(h.db, highlights, viewerDID)
	authBooks, _ := hydrateBookmarks(h.db, bookmarks, viewerDID)

	if len(collectionItems) > 0 {
		var sembleURIs []string
		for _, item := range collectionItems {
			if strings.Contains(item.AnnotationURI, "network.cosmik.card") {
				sembleURIs = append(sembleURIs, item.AnnotationURI)
			}
		}
		if len(sembleURIs) > 0 {
			ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			ensureSembleCardsIndexed(ctx, h.db, sembleURIs)
		}
	}

	authCollectionItems, _ := hydrateCollectionItems(h.db, collectionItems, viewerDID)

	collectionItemURIs := make(map[string]string)
	for _, ci := range authCollectionItems {
		var annotationURI string
		if ci.Annotation != nil {
			annotationURI = ci.Annotation.ID
		} else if ci.Highlight != nil {
			annotationURI = ci.Highlight.ID
		} else if ci.Bookmark != nil {
			annotationURI = ci.Bookmark.ID
		}
		if annotationURI != "" {
			collectionItemURIs[annotationURI] = ci.Author.DID
		}
	}

	var feed []interface{}
	for _, a := range authAnnos {
		if addedBy, exists := collectionItemURIs[a.ID]; exists && addedBy == a.Author.DID {
			continue
		}
		feed = append(feed, a)
	}
	for _, h := range authHighs {
		if addedBy, exists := collectionItemURIs[h.ID]; exists && addedBy == h.Author.DID {
			continue
		}
		feed = append(feed, h)
	}
	for _, b := range authBooks {
		if addedBy, exists := collectionItemURIs[b.ID]; exists && addedBy == b.Author.DID {
			continue
		}
		feed = append(feed, b)
	}
	for _, ci := range authCollectionItems {
		feed = append(feed, ci)
	}

	if feedType != "" && feedType != "all" && feedType != "my-feed" {
		var filtered []interface{}
		for _, item := range feed {
			isSemble := false
			var uri string
			switch v := item.(type) {
			case APIAnnotation:
				uri = v.ID
			case APIHighlight:
				uri = v.ID
			case APIBookmark:
				uri = v.ID
			case APICollectionItem:
				if v.Annotation != nil {
					uri = v.Annotation.ID
				} else if v.Highlight != nil {
					uri = v.Highlight.ID
				} else if v.Bookmark != nil {
					uri = v.Bookmark.ID
				} else {
					uri = v.ID
				}
			}
			if strings.Contains(uri, "network.cosmik") {
				isSemble = true
			}

			switch feedType {
			case "semble":
				if isSemble {
					filtered = append(filtered, item)
				}
			case "margin":
				if !isSemble {
					filtered = append(filtered, item)
				}
			case "popular", "shelved":
				filtered = append(filtered, item)
			}
		}
		feed = filtered
	}

	feed = h.filterFeedByModeration(feed, viewerDID)

	switch feedType {
	case "popular":
		sortFeedByPopularity(feed)
	default:
		sortFeed(feed)
	}

	log.Printf("[DEBUG] FeedType: %s, Total Items before slice: %d", feedType, len(feed))
	if len(feed) > 0 {
		first := feed[0]
		switch v := first.(type) {
		case APIAnnotation:
			log.Printf("[DEBUG] First Item (Annotation): %s, Likes: %d, Replies: %d", v.ID, v.LikeCount, v.ReplyCount)
		case APIHighlight:
			log.Printf("[DEBUG] First Item (Highlight): %s, Likes: %d, Replies: %d", v.ID, v.LikeCount, v.ReplyCount)
		}
	}

	if offset < len(feed) {
		feed = feed[offset:]
	} else {
		feed = []interface{}{}
	}

	if len(feed) > limit {
		feed = feed[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "Collection",
		"items":      feed,
		"totalItems": len(feed),
	})
}

func containsTag(tagsJSON *string, tag string) bool {
	if tagsJSON == nil || *tagsJSON == "" {
		return false
	}
	var tags []string
	if err := json.Unmarshal([]byte(*tagsJSON), &tags); err != nil {
		return false
	}
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func sortFeed(feed []interface{}) {
	sort.Slice(feed, func(i, j int) bool {
		t1 := getCreatedAt(feed[i])
		t2 := getCreatedAt(feed[j])
		return t1.After(t2)
	})
}

func getCreatedAt(item interface{}) time.Time {
	switch v := item.(type) {
	case APIAnnotation:
		return v.CreatedAt
	case APIHighlight:
		return v.CreatedAt
	case APIBookmark:
		return v.CreatedAt
	case APICollectionItem:
		return v.CreatedAt
	default:
		return time.Time{}
	}
}

func sortFeedByPopularity(feed []interface{}) {
	sort.Slice(feed, func(i, j int) bool {
		p1 := getPopularity(feed[i])
		p2 := getPopularity(feed[j])
		if p1 != p2 {
			return p1 > p2
		}
		t1 := getCreatedAt(feed[i])
		t2 := getCreatedAt(feed[j])
		return t1.After(t2)
	})
}

func getPopularity(item interface{}) int {
	switch v := item.(type) {
	case APIAnnotation:
		return v.LikeCount + v.ReplyCount
	case APIHighlight:
		return v.LikeCount + v.ReplyCount
	case APIBookmark:
		return v.LikeCount + v.ReplyCount
	case APICollectionItem:
		pop := 0
		if v.Annotation != nil {
			pop += v.Annotation.LikeCount + v.Annotation.ReplyCount
		}
		if v.Highlight != nil {
			pop += v.Highlight.LikeCount + v.Highlight.ReplyCount
		}
		if v.Bookmark != nil {
			pop += v.Bookmark.LikeCount + v.Bookmark.ReplyCount
		}
		return pop
	default:
		return 0
	}
}

func (h *Handler) GetAnnotation(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	serveResponse := func(data interface{}, context string) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"@context": context,
		}
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &response)
		json.NewEncoder(w).Encode(response)
	}

	if annotation, err := h.db.GetAnnotationByURI(uri); err == nil {
		if annotation.CID == nil || *annotation.CID == "" {
			parts := parseATURI(uri)
			if len(parts) >= 3 {
				did := parts[0]
				collection := parts[1]
				rkey := parts[2]

				session, err := h.refresher.GetSessionWithAutoRefresh(r)
				if err == nil {
					_ = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, _ string) error {
						record, getErr := client.GetRecord(r.Context(), did, collection, rkey)
						if getErr == nil {
							h.db.UpdateAnnotation(uri, *annotation.BodyValue, *annotation.TagsJSON, record.CID)
							cid := record.CID
							annotation.CID = &cid
						}
						return nil
					})
				}
			}
		}

		if enriched, _ := hydrateAnnotations(h.db, []db.Annotation{*annotation}, h.getViewerDID(r)); len(enriched) > 0 {
			serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
			return
		}
	}

	if highlight, err := h.db.GetHighlightByURI(uri); err == nil {
		if highlight.CID == nil || *highlight.CID == "" {
			parts := parseATURI(uri)
			if len(parts) >= 3 {
				did := parts[0]
				collection := parts[1]
				rkey := parts[2]

				session, err := h.refresher.GetSessionWithAutoRefresh(r)
				if err == nil {
					_ = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, _ string) error {
						record, getErr := client.GetRecord(r.Context(), did, collection, rkey)
						if getErr == nil {
							tagsJSON := ""
							if highlight.TagsJSON != nil {
								tagsJSON = *highlight.TagsJSON
							}
							color := ""
							if highlight.Color != nil {
								color = *highlight.Color
							}
							h.db.UpdateHighlight(uri, color, tagsJSON, record.CID)
							cid := record.CID
							highlight.CID = &cid
						}
						return nil
					})
				}
			}
		}

		if enriched, _ := hydrateHighlights(h.db, []db.Highlight{*highlight}, h.getViewerDID(r)); len(enriched) > 0 {
			serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
			return
		}
	}

	if strings.Contains(uri, "at.margin.annotation") {
		highlightURI := strings.Replace(uri, "at.margin.annotation", "at.margin.highlight", 1)
		if highlight, err := h.db.GetHighlightByURI(highlightURI); err == nil {
			if enriched, _ := hydrateHighlights(h.db, []db.Highlight{*highlight}, h.getViewerDID(r)); len(enriched) > 0 {
				serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
				return
			}
		}
	}

	if strings.Contains(uri, "at.margin.annotation") || strings.Contains(uri, "at.margin.bookmark") {
		if strings.HasPrefix(uri, "at://") {
			uriWithoutScheme := strings.TrimPrefix(uri, "at://")
			parts := strings.Split(uriWithoutScheme, "/")
			if len(parts) >= 3 {
				did := parts[0]
				rkey := parts[len(parts)-1]

				sembleURI := fmt.Sprintf("at://%s/network.cosmik.card/%s", did, rkey)

				if annotation, err := h.db.GetAnnotationByURI(sembleURI); err == nil {
					if enriched, _ := hydrateAnnotations(h.db, []db.Annotation{*annotation}, h.getViewerDID(r)); len(enriched) > 0 {
						serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
						return
					}
				}

				if bookmark, err := h.db.GetBookmarkByURI(sembleURI); err == nil {
					if enriched, _ := hydrateBookmarks(h.db, []db.Bookmark{*bookmark}, h.getViewerDID(r)); len(enriched) > 0 {
						serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
						return
					}
				}
			}
		}
	}

	if bookmark, err := h.db.GetBookmarkByURI(uri); err == nil {
		if bookmark.CID == nil || *bookmark.CID == "" {
			parts := parseATURI(uri)
			if len(parts) >= 3 {
				did := parts[0]
				collection := parts[1]
				rkey := parts[2]

				session, err := h.refresher.GetSessionWithAutoRefresh(r)
				if err == nil {
					_ = h.refresher.ExecuteWithAutoRefresh(r, session, func(client *xrpc.Client, _ string) error {
						record, getErr := client.GetRecord(r.Context(), did, collection, rkey)
						if getErr == nil {
							tagsJSON := ""
							if bookmark.TagsJSON != nil {
								tagsJSON = *bookmark.TagsJSON
							}
							title := ""
							if bookmark.Title != nil {
								title = *bookmark.Title
							}
							desc := ""
							if bookmark.Description != nil {
								desc = *bookmark.Description
							}
							h.db.UpdateBookmark(uri, title, desc, tagsJSON, record.CID)
							cid := record.CID
							bookmark.CID = &cid
						}
						return nil
					})
				}
			}
		}

		if enriched, _ := hydrateBookmarks(h.db, []db.Bookmark{*bookmark}, h.getViewerDID(r)); len(enriched) > 0 {
			serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
			return
		}
	}

	if strings.Contains(uri, "at.margin.annotation") {
		bookmarkURI := strings.Replace(uri, "at.margin.annotation", "at.margin.bookmark", 1)
		if bookmark, err := h.db.GetBookmarkByURI(bookmarkURI); err == nil {
			if enriched, _ := hydrateBookmarks(h.db, []db.Bookmark{*bookmark}, h.getViewerDID(r)); len(enriched) > 0 {
				serveResponse(enriched[0], "http://www.w3.org/ns/anno.jsonld")
				return
			}
		}
	}

	http.Error(w, "Annotation, Highlight, or Bookmark not found", http.StatusNotFound)

}

func (h *Handler) GetByTarget(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		source = r.URL.Query().Get("url")
	}
	if source == "" {
		http.Error(w, "source or url parameter required", http.StatusBadRequest)
		return
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	urlHash := db.HashURL(source)
	rawHash := db.HashString(source)

	annotations, _ := h.db.GetAnnotationsByTargetHash(urlHash, limit, offset)
	highlights, _ := h.db.GetHighlightsByTargetHash(urlHash, limit, offset)
	bookmarks, _ := h.db.GetBookmarksByTargetHash(urlHash, limit, offset)

	if rawHash != urlHash {
		rawAnnotations, _ := h.db.GetAnnotationsByTargetHash(rawHash, limit, offset)
		rawHighlights, _ := h.db.GetHighlightsByTargetHash(rawHash, limit, offset)
		rawBookmarks, _ := h.db.GetBookmarksByTargetHash(rawHash, limit, offset)

		annotations = mergeAnnotations(annotations, rawAnnotations)
		highlights = mergeHighlights(highlights, rawHighlights)
		bookmarks = mergeBookmarks(bookmarks, rawBookmarks)
	}

	enrichedAnnotations, _ := hydrateAnnotations(h.db, annotations, h.getViewerDID(r))
	enrichedHighlights, _ := hydrateHighlights(h.db, highlights, h.getViewerDID(r))
	enrichedBookmarks, _ := hydrateBookmarks(h.db, bookmarks, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":    "http://www.w3.org/ns/anno.jsonld",
		"source":      source,
		"sourceHash":  urlHash,
		"annotations": enrichedAnnotations,
		"highlights":  enrichedHighlights,
		"bookmarks":   enrichedBookmarks,
	})
}

func (h *Handler) DiscoverForURL(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		source = r.URL.Query().Get("url")
	}
	if source == "" {
		http.Error(w, "source or url parameter required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	annotations, highlights, bookmarks, err := ConstellationClient.GetAllItemsForURL(ctx, source)
	if err != nil {
		log.Printf("Constellation discover error, falling back to local: %v", err)
		h.GetByTarget(w, r)
		return
	}

	var annotationURIs, highlightURIs, bookmarkURIs []string
	seenURIs := make(map[string]bool)

	for _, link := range annotations {
		if !seenURIs[link.URI] {
			annotationURIs = append(annotationURIs, link.URI)
			seenURIs[link.URI] = true
		}
	}
	for _, link := range highlights {
		if !seenURIs[link.URI] {
			highlightURIs = append(highlightURIs, link.URI)
			seenURIs[link.URI] = true
		}
	}
	for _, link := range bookmarks {
		if !seenURIs[link.URI] {
			bookmarkURIs = append(bookmarkURIs, link.URI)
			seenURIs[link.URI] = true
		}
	}

	localAnnotations, _ := h.db.GetAnnotationsByURIs(annotationURIs)
	localHighlights, _ := h.db.GetHighlightsByURIs(highlightURIs)
	localBookmarks, _ := h.db.GetBookmarksByURIs(bookmarkURIs)

	urlHash := db.HashURL(source)
	dbAnnotations, _ := h.db.GetAnnotationsByTargetHash(urlHash, 100, 0)
	dbHighlights, _ := h.db.GetHighlightsByTargetHash(urlHash, 100, 0)
	dbBookmarks, _ := h.db.GetBookmarksByTargetHash(urlHash, 100, 0)

	annoMap := make(map[string]db.Annotation)
	for _, a := range localAnnotations {
		annoMap[a.URI] = a
	}
	for _, a := range dbAnnotations {
		annoMap[a.URI] = a
	}

	highMap := make(map[string]db.Highlight)
	for _, h := range localHighlights {
		highMap[h.URI] = h
	}
	for _, h := range dbHighlights {
		highMap[h.URI] = h
	}

	bookMap := make(map[string]db.Bookmark)
	for _, b := range localBookmarks {
		bookMap[b.URI] = b
	}
	for _, b := range dbBookmarks {
		bookMap[b.URI] = b
	}

	var mergedAnnotations []db.Annotation
	for _, a := range annoMap {
		mergedAnnotations = append(mergedAnnotations, a)
	}
	var mergedHighlights []db.Highlight
	for _, h := range highMap {
		mergedHighlights = append(mergedHighlights, h)
	}
	var mergedBookmarks []db.Bookmark
	for _, b := range bookMap {
		mergedBookmarks = append(mergedBookmarks, b)
	}

	viewerDID := h.getViewerDID(r)
	enrichedAnnotations, _ := hydrateAnnotations(h.db, mergedAnnotations, viewerDID)
	enrichedHighlights, _ := hydrateHighlights(h.db, mergedHighlights, viewerDID)
	enrichedBookmarks, _ := hydrateBookmarks(h.db, mergedBookmarks, viewerDID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":          "http://www.w3.org/ns/anno.jsonld",
		"source":            source,
		"sourceHash":        urlHash,
		"annotations":       enrichedAnnotations,
		"highlights":        enrichedHighlights,
		"bookmarks":         enrichedBookmarks,
		"networkDiscovered": len(annotations) + len(highlights) + len(bookmarks),
	})
}

func (h *Handler) GetHighlights(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("creator")
	tag := r.URL.Query().Get("tag")
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	var highlights []db.Highlight
	var err error

	if did != "" {
		highlights, err = h.db.GetHighlightsByAuthor(did, limit, offset)
	} else if tag != "" {
		highlights, err = h.db.GetHighlightsByTag(tag, limit, offset)
	} else {
		highlights, err = h.db.GetRecentHighlights(limit, offset)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateHighlights(h.db, highlights, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "HighlightCollection",
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("creator")
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	if did == "" {
		http.Error(w, "creator parameter required", http.StatusBadRequest)
		return
	}

	bookmarks, err := h.db.GetBookmarksByAuthor(did, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateBookmarks(h.db, bookmarks, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "BookmarkCollection",
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetUserAnnotations(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	var annotations []db.Annotation
	var err error

	viewerDID := h.getViewerDID(r)

	if offset == 0 && viewerDID != "" && did == viewerDID {
		go func() {
			if _, err := h.FetchLatestUserRecords(r, did, xrpc.CollectionAnnotation, limit); err != nil {
				log.Printf("Background sync error (annotations): %v", err)
			}
		}()
	}

	annotations, err = h.db.GetAnnotationsByAuthor(did, limit, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateAnnotations(h.db, annotations, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "AnnotationCollection",
		"creator":    did,
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetUserHighlights(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	var highlights []db.Highlight
	var err error

	viewerDID := h.getViewerDID(r)

	if offset == 0 && viewerDID != "" && did == viewerDID {
		go func() {
			if _, err := h.FetchLatestUserRecords(r, did, xrpc.CollectionHighlight, limit); err != nil {
				log.Printf("Background sync error (highlights): %v", err)
			}
		}()
	}

	highlights, err = h.db.GetHighlightsByAuthor(did, limit, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateHighlights(h.db, highlights, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "HighlightCollection",
		"creator":    did,
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetUserBookmarks(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	var bookmarks []db.Bookmark
	var err error

	viewerDID := h.getViewerDID(r)

	if offset == 0 && viewerDID != "" && did == viewerDID {
		go func() {
			if _, err := h.FetchLatestUserRecords(r, did, xrpc.CollectionBookmark, limit); err != nil {
				log.Printf("Background sync error (bookmarks): %v", err)
			}
		}()
	}

	bookmarks, err = h.db.GetBookmarksByAuthor(did, limit, offset)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateBookmarks(h.db, bookmarks, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "BookmarkCollection",
		"creator":    did,
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetUserTargetItems(w http.ResponseWriter, r *http.Request) {
	did := chi.URLParam(r, "did")
	if decoded, err := url.QueryUnescape(did); err == nil {
		did = decoded
	}

	source := r.URL.Query().Get("source")
	if source == "" {
		source = r.URL.Query().Get("url")
	}
	if source == "" {
		http.Error(w, "source or url parameter required", http.StatusBadRequest)
		return
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	urlHash := db.HashURL(source)

	annotations, _ := h.db.GetAnnotationsByAuthorAndTargetHash(did, urlHash, limit, offset)
	highlights, _ := h.db.GetHighlightsByAuthorAndTargetHash(did, urlHash, limit, offset)

	enrichedAnnotations, _ := hydrateAnnotations(h.db, annotations, h.getViewerDID(r))
	enrichedHighlights, _ := hydrateHighlights(h.db, highlights, h.getViewerDID(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":    "http://www.w3.org/ns/anno.jsonld",
		"creator":     did,
		"source":      source,
		"sourceHash":  urlHash,
		"annotations": enrichedAnnotations,
		"highlights":  enrichedHighlights,
	})
}

func (h *Handler) GetReplies(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	replies, err := h.db.GetRepliesByRoot(uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	enriched, _ := hydrateReplies(h.db, replies)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"@context":   "http://www.w3.org/ns/anno.jsonld",
		"type":       "ReplyCollection",
		"inReplyTo":  uri,
		"items":      enriched,
		"totalItems": len(enriched),
	})
}

func (h *Handler) GetLikeCount(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	count, err := h.db.GetLikeCount(uri)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	liked := false
	cookie, err := r.Cookie("margin_session")
	if err == nil && cookie != nil {
		session, err := h.refresher.GetSessionWithAutoRefresh(r)
		if err == nil {
			userLike, err := h.db.GetLikeByUserAndSubject(session.DID, uri)
			if err == nil && userLike != nil {
				liked = true
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": count,
		"liked": liked,
	})
}

func (h *Handler) GetEditHistory(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Query().Get("uri")
	if uri == "" {
		http.Error(w, "uri query parameter required", http.StatusBadRequest)
		return
	}

	history, err := h.db.GetEditHistory(uri)
	if err != nil {
		http.Error(w, "Failed to fetch edit history", http.StatusInternalServerError)
		return
	}

	if history == nil {
		history = []db.EditHistory{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func parseIntParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}

func (h *Handler) GetURLMetadata(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		http.Error(w, "url parameter required", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"title": "", "error": "failed to fetch"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 500*1024))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"title": ""})
		return
	}

	content := string(body)

	extract := func(key string) string {
		attr := fmt.Sprintf("property=\"og:%s\"", key)
		if idx := strings.Index(content, attr); idx != -1 {
			rest := content[idx:]
			if contentIdx := strings.Index(rest, "content=\""); contentIdx != -1 {
				start := contentIdx + 9
				if end := strings.Index(rest[start:], "\""); end != -1 {
					return rest[start : start+end]
				}
			}
		}

		attr = fmt.Sprintf("name=\"%s\"", key)
		if idx := strings.Index(content, attr); idx != -1 {
			rest := content[idx:]
			if contentIdx := strings.Index(rest, "content=\""); contentIdx != -1 {
				start := contentIdx + 9
				if end := strings.Index(rest[start:], "\""); end != -1 {
					return rest[start : start+end]
				}
			}
		}
		return ""
	}

	title := extract("title")
	if title == "" {
		if idx := strings.Index(content, "<title>"); idx != -1 {
			start := idx + 7
			if end := strings.Index(content[start:], "</title>"); end != -1 {
				title = content[start : start+end]
			}
		}
	}

	description := extract("description")
	image := extract("image")

	var favicon string
	findIcon := func(rel string) string {
		search := fmt.Sprintf("rel=\"%s\"", rel)
		if idx := strings.Index(content, search); idx != -1 {
			startTag := strings.LastIndex(content[:idx], "<link")
			if startTag != -1 {
				endTag := strings.Index(content[startTag:], ">")
				if endTag != -1 {
					tag := content[startTag : startTag+endTag]
					if hrefIdx := strings.Index(tag, "href=\""); hrefIdx != -1 {
						start := hrefIdx + 6
						if end := strings.Index(tag[start:], "\""); end != -1 {
							return tag[start : start+end]
						}
					}
				}
			}
		}
		return ""
	}

	favicon = findIcon("icon")
	if favicon == "" {
		favicon = findIcon("shortcut icon")
	}
	if favicon == "" {
		favicon = findIcon("apple-touch-icon")
	}

	resolveURL := func(base, target string) string {
		if target == "" {
			return ""
		}
		if strings.HasPrefix(target, "http") {
			return target
		}
		if strings.HasPrefix(target, "//") {
			return "https:" + target
		}
		u, err := url.Parse(base)
		if err != nil {
			return target
		}
		t, err := url.Parse(target)
		if err != nil {
			return target
		}
		return u.ResolveReference(t).String()
	}

	image = resolveURL(targetURL, image)
	favicon = resolveURL(targetURL, favicon)

	if favicon == "" {
		u, err := url.Parse(targetURL)
		if err == nil {
			favicon = fmt.Sprintf("%s://%s/favicon.ico", u.Scheme, u.Host)
		}
	}

	data := map[string]string{
		"title":       title,
		"description": description,
		"image":       image,
		"icon":        favicon,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	notifications, err := h.db.GetNotifications(session.DID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	enriched, err := hydrateNotifications(h.db, notifications)
	if err != nil {
		log.Printf("Failed to hydrate notifications: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	if enriched != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"items": enriched})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{"items": notifications})
	}
}

func (h *Handler) GetUnreadNotificationCount(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	count, err := h.db.GetUnreadNotificationCount(session.DID)
	if err != nil {
		http.Error(w, "Failed to get count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": count})
}

func (h *Handler) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	session, err := h.refresher.GetSessionWithAutoRefresh(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := h.db.MarkNotificationsRead(session.DID); err != nil {
		http.Error(w, "Failed to mark as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
func (h *Handler) getViewerDID(r *http.Request) string {
	cookie, err := r.Cookie("margin_session")
	if err != nil {
		return ""
	}
	did, _, _, _, _, err := h.db.GetSession(cookie.Value)
	if err != nil {
		return ""
	}
	return did
}

func getItemAuthorDID(item interface{}) string {
	switch v := item.(type) {
	case APIAnnotation:
		return v.Author.DID
	case APIHighlight:
		return v.Author.DID
	case APIBookmark:
		return v.Author.DID
	case APICollectionItem:
		return v.Author.DID
	default:
		return ""
	}
}

func (h *Handler) filterFeedByModeration(feed []interface{}, viewerDID string) []interface{} {
	if viewerDID == "" {
		return feed
	}

	hiddenDIDs, err := h.db.GetAllHiddenDIDs(viewerDID)
	if err != nil || len(hiddenDIDs) == 0 {
		return feed
	}

	var filtered []interface{}
	for _, item := range feed {
		authorDID := getItemAuthorDID(item)
		if authorDID != "" && hiddenDIDs[authorDID] {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func mergeAnnotations(a, b []db.Annotation) []db.Annotation {
	seen := make(map[string]bool)
	var result []db.Annotation
	for _, item := range a {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	for _, item := range b {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	return result
}

func mergeHighlights(a, b []db.Highlight) []db.Highlight {
	seen := make(map[string]bool)
	var result []db.Highlight
	for _, item := range a {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	for _, item := range b {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	return result
}

func mergeBookmarks(a, b []db.Bookmark) []db.Bookmark {
	seen := make(map[string]bool)
	var result []db.Bookmark
	for _, item := range a {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	for _, item := range b {
		if !seen[item.URI] {
			seen[item.URI] = true
			result = append(result, item)
		}
	}
	return result
}
