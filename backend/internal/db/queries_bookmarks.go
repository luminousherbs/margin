package db

import (
	"time"
)

func (db *DB) CreateBookmark(b *Bookmark) error {
	_, err := db.Exec(`
		INSERT INTO bookmarks (uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT(uri) DO UPDATE SET
			source = EXCLUDED.source,
			source_hash = EXCLUDED.source_hash,
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			tags_json = EXCLUDED.tags_json,
			indexed_at = EXCLUDED.indexed_at,
			cid = EXCLUDED.cid
	`, b.URI, b.AuthorDID, b.Source, b.SourceHash, b.Title, b.Description, b.TagsJSON, b.CreatedAt, b.IndexedAt, b.CID)
	return err
}

func (db *DB) GetBookmarkByURI(uri string) (*Bookmark, error) {
	var b Bookmark
	err := db.QueryRow(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri = $1
	`, uri).Scan(&b.URI, &b.AuthorDID, &b.Source, &b.SourceHash, &b.Title, &b.Description, &b.TagsJSON, &b.CreatedAt, &b.IndexedAt, &b.CID)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (db *DB) GetRecentBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetPopularBookmarks(limit, offset int) ([]Bookmark, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			b.uri, b.author_did, b.source, b.source_hash, b.title,
			b.description, b.tags_json, b.created_at, b.indexed_at, b.cid
		FROM bookmarks b
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM likes WHERE subject_uri = b.uri
		) l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM replies WHERE root_uri = b.uri
		) r ON true
		WHERE b.created_at > $1 AND (l.cnt + r.cnt) > 0
		ORDER BY (l.cnt + r.cnt) DESC, b.created_at DESC
		LIMIT $2 OFFSET $3
	`, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetShelvedBookmarks(limit, offset int) ([]Bookmark, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			b.uri, b.author_did, b.source, b.source_hash, b.title,
			b.description, b.tags_json, b.created_at, b.indexed_at, b.cid
		FROM bookmarks b
		WHERE b.created_at < $1 AND b.created_at > $2
			AND NOT EXISTS (SELECT 1 FROM likes WHERE subject_uri = b.uri)
			AND NOT EXISTS (SELECT 1 FROM replies WHERE root_uri = b.uri)
		ORDER BY RANDOM()
		LIMIT $3 OFFSET $4
	`, olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetMarginBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetSembleBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetMarginBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetSembleBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetMarginBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetSembleBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetMarginBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1 AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetSembleBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = $1 AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) DeleteBookmark(uri string) error {
	_, err := db.Exec(`DELETE FROM bookmarks WHERE uri = $1`, uri)
	return err
}

func (db *DB) UpdateBookmark(uri, title, description, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE bookmarks
		SET title = $1, description = $2, tags_json = $3, cid = $4, indexed_at = $5
		WHERE uri = $6
	`, title, description, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) GetBookmarksByURIs(uris []string) ([]Bookmark, error) {
	if len(uris) == 0 {
		return []Bookmark{}, nil
	}

	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri = ANY($1)
	`, pqStringArray(uris))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func (db *DB) GetBookmarkURIs(authorDID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT uri FROM bookmarks WHERE author_did = $1
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

func (db *DB) GetBookmarksByTargetHash(targetHash string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE source_hash = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBookmarks(rows)
}

func scanBookmarks(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]Bookmark, error) {
	var bookmarks []Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.URI, &b.AuthorDID, &b.Source, &b.SourceHash, &b.Title, &b.Description, &b.TagsJSON, &b.CreatedAt, &b.IndexedAt, &b.CID); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, b)
	}
	return bookmarks, nil
}
