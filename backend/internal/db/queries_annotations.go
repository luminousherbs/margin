package db

import (
	"time"
)

func (db *DB) CreateAnnotation(a *Annotation) error {
	_, err := db.Exec(db.Rebind(`
		INSERT INTO annotations (uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			motivation = excluded.motivation,
			body_value = excluded.body_value,
			body_format = excluded.body_format,
			body_uri = excluded.body_uri,
			target_source = excluded.target_source,
			target_hash = excluded.target_hash,
			target_title = excluded.target_title,
			selector_json = excluded.selector_json,
			tags_json = excluded.tags_json,
			indexed_at = excluded.indexed_at,
			cid = excluded.cid
	`), a.URI, a.AuthorDID, a.Motivation, a.BodyValue, a.BodyFormat, a.BodyURI, a.TargetSource, a.TargetHash, a.TargetTitle, a.SelectorJSON, a.TagsJSON, a.CreatedAt, a.IndexedAt, a.CID)
	return err
}

func (db *DB) GetAnnotationByURI(uri string) (*Annotation, error) {
	var a Annotation
	err := db.QueryRow(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE uri = ?
	`), uri).Scan(&a.URI, &a.AuthorDID, &a.Motivation, &a.BodyValue, &a.BodyFormat, &a.BodyURI, &a.TargetSource, &a.TargetHash, &a.TargetTitle, &a.SelectorJSON, &a.TagsJSON, &a.CreatedAt, &a.IndexedAt, &a.CID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) GetAnnotationsByTargetHash(targetHash string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE target_hash = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByMotivation(motivation string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE motivation = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), motivation, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetRecentAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetPopularAnnotations(limit, offset int) ([]Annotation, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			a.uri, a.author_did, a.motivation, a.body_value, a.body_format, 
			a.body_uri, a.target_source, a.target_hash, a.target_title, 
			a.selector_json, a.tags_json, a.created_at, a.indexed_at, a.cid
		FROM annotations a
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = a.uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = a.uri
		WHERE a.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) > 0
		ORDER BY (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) DESC, a.created_at DESC
		LIMIT ? OFFSET ?
	`), since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetShelvedAnnotations(limit, offset int) ([]Annotation, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			a.uri, a.author_did, a.motivation, a.body_value, a.body_format, 
			a.body_uri, a.target_source, a.target_hash, a.target_title, 
			a.selector_json, a.tags_json, a.created_at, a.indexed_at, a.cid
		FROM annotations a
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = a.uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = a.uri
		WHERE a.created_at < ? AND a.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) = 0
		ORDER BY RANDOM()
		LIMIT ? OFFSET ?
	`), olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE tags_json LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE tags_json LIKE ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE tags_json LIKE ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) DeleteAnnotation(uri string) error {
	_, err := db.Exec(db.Rebind(`DELETE FROM annotations WHERE uri = ?`), uri)
	return err
}

func (db *DB) UpdateAnnotation(uri, bodyValue, tagsJSON, cid string) error {
	_, err := db.Exec(db.Rebind(`
		UPDATE annotations 
		SET body_value = ?, tags_json = ?, cid = ?, indexed_at = ?
		WHERE uri = ?
	`), bodyValue, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) GetAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND tags_json LIKE ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND tags_json LIKE ? AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	pattern := "%\"" + tag + "\"%"
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND tags_json LIKE ? AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, pattern, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByAuthorAndTargetHash(authorDID, targetHash string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
		WHERE author_did = ? AND target_hash = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), authorDID, targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByURIs(uris []string) ([]Annotation, error) {
	if len(uris) == 0 {
		return []Annotation{}, nil
	}

	query := db.Rebind(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM annotations
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

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationURIs(authorDID string) ([]string, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri FROM annotations WHERE author_did = ?
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
