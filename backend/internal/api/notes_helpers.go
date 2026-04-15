package api

import (
	"context"
	"encoding/json"
	"time"

	"margin.at/internal/domain"
	"margin.at/internal/xrpc"
)

var validSelfLabelSet = map[string]bool{
	"sexual":     true,
	"nudity":     true,
	"violence":   true,
	"gore":       true,
	"spam":       true,
	"misleading": true,
}

func filterSelfLabels(labels []string) []string {
	if len(labels) == 0 {
		return nil
	}
	out := make([]string, 0, len(labels))
	for _, l := range labels {
		if validSelfLabelSet[l] {
			out = append(out, l)
		}
	}
	return out
}

func putRecordWithRetry(
	ctx context.Context,
	client *xrpc.Client,
	did, collection, rkey string,
	record interface{ Validate() error },
) (*xrpc.PutRecordOutput, error) {
	if err := record.Validate(); err != nil {
		return nil, err
	}
	result, err := client.PutRecord(ctx, did, collection, rkey, record)
	if err != nil {
		_ = client.DeleteRecord(ctx, did, collection, rkey)
		result, err = client.PutRecord(ctx, did, collection, rkey, record)
	}
	return result, err
}

func (s *NoteWriteService) checkDuplicateAnnotation(did, url, text string) (*domain.Annotation, error) {
	recent, err := s.db.GetAnnotationsByAuthor(did, 5, 0)
	if err != nil {
		return nil, err
	}
	for i := range recent {
		a := &recent[i]
		if a.TargetSource == url &&
			((a.BodyValue == nil && text == "") || (a.BodyValue != nil && *a.BodyValue == text)) &&
			time.Since(a.CreatedAt) < 10*time.Second {
			return a, nil
		}
	}
	return nil, nil
}

func (s *NoteWriteService) checkDuplicateHighlight(did, url string, selector json.RawMessage) (*domain.Highlight, error) {
	recent, err := s.db.GetHighlightsByAuthor(did, 5, 0)
	if err != nil {
		return nil, err
	}
	for i := range recent {
		h := &recent[i]
		if h.TargetSource != url || time.Since(h.CreatedAt) >= 10*time.Second {
			continue
		}
		if selector == nil && h.SelectorJSON == nil {
			return h, nil
		}
		if selector != nil && h.SelectorJSON != nil {
			b, _ := json.Marshal(selector)
			if *h.SelectorJSON == string(b) {
				return h, nil
			}
		}
	}
	return nil, nil
}

func (s *NoteWriteService) checkDuplicateBookmark(did, url string) (*domain.Bookmark, error) {
	urlHash := s.db.HashURL(url)
	bookmarks, err := s.db.GetBookmarksByTargetHash(urlHash, 50, 0)
	if err != nil {
		return nil, err
	}
	for i := range bookmarks {
		b := &bookmarks[i]
		if b.AuthorDID == did && b.Source == url {
			return b, nil
		}
	}
	return nil, nil
}
