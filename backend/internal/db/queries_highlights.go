package db

import (
	"time"
)

func (db *DB) CreateHighlight(h *Highlight) error {
	if taken, _ := db.IsTakenDown(h.URI); taken {
		return nil
	}
	_, err := db.Exec(`
		INSERT INTO highlights (uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT(uri) DO UPDATE SET
			target_source = EXCLUDED.target_source,
			target_hash = EXCLUDED.target_hash,
			target_title = EXCLUDED.target_title,
			selector_json = EXCLUDED.selector_json,
			color = EXCLUDED.color,
			tags_json = EXCLUDED.tags_json,
			indexed_at = EXCLUDED.indexed_at,
			cid = EXCLUDED.cid
	`, h.URI, h.AuthorDID, h.TargetSource, h.TargetHash, h.TargetTitle, h.SelectorJSON, h.Color, h.TagsJSON, h.CreatedAt, h.IndexedAt, h.CID)
	return err
}

func (db *DB) GetHighlightByURI(uri string) (*Highlight, error) {
	var h Highlight
	err := db.QueryRow(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE uri = $1
	`, uri).Scan(&h.URI, &h.AuthorDID, &h.TargetSource, &h.TargetHash, &h.TargetTitle, &h.SelectorJSON, &h.Color, &h.TagsJSON, &h.CreatedAt, &h.IndexedAt, &h.CID)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (db *DB) GetRecentHighlights(limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetPopularHighlights(limit, offset int) ([]Highlight, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			h.uri, h.author_did, h.target_source, h.target_hash, h.target_title,
			h.selector_json, h.color, h.tags_json, h.created_at, h.indexed_at, h.cid
		FROM all_highlights h
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM likes WHERE subject_uri = h.uri
		) l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM replies WHERE root_uri = h.uri
		) r ON true
		WHERE h.created_at > $1 AND (l.cnt + r.cnt) > 0
		ORDER BY (l.cnt + r.cnt) DESC, h.created_at DESC
		LIMIT $2 OFFSET $3
	`, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetShelvedHighlights(limit, offset int) ([]Highlight, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			h.uri, h.author_did, h.target_source, h.target_hash, h.target_title,
			h.selector_json, h.color, h.tags_json, h.created_at, h.indexed_at, h.cid
		FROM all_highlights h
		WHERE h.created_at < $1 AND h.created_at > $2
			AND NOT EXISTS (SELECT 1 FROM likes WHERE subject_uri = h.uri)
			AND NOT EXISTS (SELECT 1 FROM replies WHERE root_uri = h.uri)
		ORDER BY RANDOM()
		LIMIT $3 OFFSET $4
	`, olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetMarginHighlights(limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetSembleHighlights(limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightsByTag(tag string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetMarginHighlightsByTag(tag string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetSembleHighlightsByTag(tag string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetMarginHighlightsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetSembleHighlightsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightsByTargetHash(targetHash string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE target_hash = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightsByAuthor(authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetMarginHighlightsByAuthor(authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetSembleHighlightsByAuthor(authorDID string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightsByAuthorAndTargetHash(authorDID, targetHash string, limit, offset int) ([]Highlight, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE author_did = $1 AND target_hash = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) DeleteHighlight(uri string) error {
	_, err := db.Exec(`DELETE FROM highlights WHERE uri = $1`, uri)
	return err
}

func (db *DB) UpdateHighlight(uri, color, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE highlights
		SET color = $1, tags_json = $2, cid = $3, indexed_at = $4
		WHERE uri = $5
	`, color, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) GetHighlightsByURIs(uris []string) ([]Highlight, error) {
	if len(uris) == 0 {
		return []Highlight{}, nil
	}

	rows, err := db.Query(`
		SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
		FROM all_highlights
		WHERE uri = ANY($1)
	`, pqStringArray(uris))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHighlights(rows)
}

func (db *DB) GetHighlightURIs(authorDID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT uri FROM all_highlights WHERE author_did = $1
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uris []string
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		uris = append(uris, uri)
	}
	return uris, nil
}

func scanHighlights(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]Highlight, error) {
	var highlights []Highlight
	for rows.Next() {
		var h Highlight
		if err := rows.Scan(&h.URI, &h.AuthorDID, &h.TargetSource, &h.TargetHash, &h.TargetTitle, &h.SelectorJSON, &h.Color, &h.TagsJSON, &h.CreatedAt, &h.IndexedAt, &h.CID); err != nil {
			return nil, err
		}
		highlights = append(highlights, h)
	}
	return highlights, nil
}
