package db

import (
	"time"
)

func (db *DB) CreateBookmark(b *Bookmark) error {
	_, err := db.Exec(db.Rebind(`
		INSERT INTO bookmarks (uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			source = excluded.source,
			source_hash = excluded.source_hash,
			title = excluded.title,
			description = excluded.description,
			tags_json = excluded.tags_json,
			indexed_at = excluded.indexed_at,
			cid = excluded.cid
	`), b.URI, b.AuthorDID, b.Source, b.SourceHash, b.Title, b.Description, b.TagsJSON, b.CreatedAt, b.IndexedAt, b.CID)
	return err
}

func (db *DB) GetBookmarkByURI(uri string) (*Bookmark, error) {
	var b Bookmark
	err := db.QueryRow(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri = ?
	`), uri).Scan(&b.URI, &b.AuthorDID, &b.Source, &b.SourceHash, &b.Title, &b.Description, &b.TagsJSON, &b.CreatedAt, &b.IndexedAt, &b.CID)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (db *DB) GetRecentBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetPopularBookmarks(limit, offset int) ([]Bookmark, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			b.uri, b.author_did, b.source, b.source_hash, b.title, 
			b.description, b.tags_json, b.created_at, b.indexed_at, b.cid
		FROM bookmarks b
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = b.uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = b.uri
		WHERE b.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) > 0
		ORDER BY (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) DESC, b.created_at DESC
		LIMIT ? OFFSET ?
	`), since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetShelvedBookmarks(limit, offset int) ([]Bookmark, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			b.uri, b.author_did, b.source, b.source_hash, b.title, 
			b.description, b.tags_json, b.created_at, b.indexed_at, b.cid
		FROM bookmarks b
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = b.uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = b.uri
		WHERE b.created_at < ? AND b.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) = 0
		ORDER BY RANDOM()
		LIMIT ? OFFSET ?
	`), olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetMarginBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetSembleBookmarks(limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE tags_json LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetMarginBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE tags_json LIKE ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetSembleBookmarksByTag(tag string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE tags_json LIKE ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ? AND tags_json LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetMarginBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ? AND tags_json LIKE ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetSembleBookmarksByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Bookmark, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ? AND tags_json LIKE ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetMarginBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetSembleBookmarksByAuthor(authorDID string, limit, offset int) ([]Bookmark, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE author_did = ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) DeleteBookmark(uri string) error {
	_, err := db.Exec(db.Rebind(`DELETE FROM bookmarks WHERE uri = ?`), uri)
	return err
}

func (db *DB) UpdateBookmark(uri, title, description, tagsJSON, cid string) error {
	_, err := db.Exec(db.Rebind(`
		UPDATE bookmarks 
		SET title = ?, description = ?, tags_json = ?, cid = ?, indexed_at = ?
		WHERE uri = ?
	`), title, description, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) GetBookmarksByURIs(uris []string) ([]Bookmark, error) {
	if len(uris) == 0 {
		return []Bookmark{}, nil
	}

	query := db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE uri IN (` + buildPlaceholders(len(uris)) + `)
	`)

	args := make([]interface{}, len(uris))
	for i, uri := range uris {
		args[i] = uri
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) GetBookmarkURIs(authorDID string) ([]string, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri FROM bookmarks WHERE author_did = ?
	`), authorDID)
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
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
		FROM bookmarks
		WHERE source_hash = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
