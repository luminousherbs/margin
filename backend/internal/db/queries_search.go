package db

func (db *DB) SearchAnnotations(query string, authorDID string, limit, offset int) ([]Annotation, error) {
	pattern := "%" + query + "%"

	var baseQuery string
	var args []interface{}

	if authorDID != "" {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
			FROM annotations
			WHERE author_did = ?
			  AND (body_value LIKE ? OR target_source LIKE ? OR target_title LIKE ? OR tags_json LIKE ? OR selector_json LIKE ?)
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{authorDID, pattern, pattern, pattern, pattern, pattern, limit, offset}
	} else {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
			FROM annotations
			WHERE body_value LIKE ? OR target_source LIKE ? OR target_title LIKE ? OR tags_json LIKE ? OR selector_json LIKE ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{pattern, pattern, pattern, pattern, pattern, limit, offset}
	}

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) SearchHighlights(query string, authorDID string, limit, offset int) ([]Highlight, error) {
	pattern := "%" + query + "%"

	var baseQuery string
	var args []interface{}

	if authorDID != "" {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
			FROM highlights
			WHERE author_did = ?
			  AND (target_source LIKE ? OR target_title LIKE ? OR selector_json LIKE ? OR tags_json LIKE ?)
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{authorDID, pattern, pattern, pattern, pattern, limit, offset}
	} else {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
			FROM highlights
			WHERE target_source LIKE ? OR target_title LIKE ? OR selector_json LIKE ? OR tags_json LIKE ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{pattern, pattern, pattern, pattern, limit, offset}
	}

	rows, err := db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

func (db *DB) SearchBookmarks(query string, authorDID string, limit, offset int) ([]Bookmark, error) {
	pattern := "%" + query + "%"

	var baseQuery string
	var args []interface{}

	if authorDID != "" {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
			FROM bookmarks
			WHERE author_did = ?
			  AND (source LIKE ? OR title LIKE ? OR description LIKE ? OR tags_json LIKE ?)
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{authorDID, pattern, pattern, pattern, pattern, limit, offset}
	} else {
		baseQuery = db.Rebind(`
			SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
			FROM bookmarks
			WHERE source LIKE ? OR title LIKE ? OR description LIKE ? OR tags_json LIKE ?
			ORDER BY created_at DESC
			LIMIT ? OFFSET ?
		`)
		args = []interface{}{pattern, pattern, pattern, pattern, limit, offset}
	}

	rows, err := db.Query(baseQuery, args...)
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
