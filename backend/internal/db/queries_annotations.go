package db

import (
	"time"
)

func (db *DB) CreateAnnotation(a *Annotation) error {
	if taken, _ := db.IsTakenDown(a.URI); taken {
		return nil
	}
	_, err := db.Exec(`
		INSERT INTO annotations (uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT(uri) DO UPDATE SET
			motivation = EXCLUDED.motivation,
			body_value = EXCLUDED.body_value,
			body_format = EXCLUDED.body_format,
			body_uri = EXCLUDED.body_uri,
			target_source = EXCLUDED.target_source,
			target_hash = EXCLUDED.target_hash,
			target_title = EXCLUDED.target_title,
			selector_json = EXCLUDED.selector_json,
			tags_json = EXCLUDED.tags_json,
			indexed_at = EXCLUDED.indexed_at,
			cid = EXCLUDED.cid
	`, a.URI, a.AuthorDID, a.Motivation, a.BodyValue, a.BodyFormat, a.BodyURI, a.TargetSource, a.TargetHash, a.TargetTitle, a.SelectorJSON, a.TagsJSON, a.CreatedAt, a.IndexedAt, a.CID)
	return err
}

func (db *DB) GetAnnotationByURI(uri string) (*Annotation, error) {
	var a Annotation
	err := db.QueryRow(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE uri = $1
	`, uri).Scan(&a.URI, &a.AuthorDID, &a.Motivation, &a.BodyValue, &a.BodyFormat, &a.BodyURI, &a.TargetSource, &a.TargetHash, &a.TargetTitle, &a.SelectorJSON, &a.TagsJSON, &a.CreatedAt, &a.IndexedAt, &a.CID)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (db *DB) GetAnnotationsByTargetHash(targetHash string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE target_hash = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, targetHash, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByAuthor(authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, authorDID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByMotivation(motivation string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE motivation = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, motivation, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetRecentAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetPopularAnnotations(limit, offset int) ([]Annotation, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			a.uri, a.author_did, a.motivation, a.body_value, a.body_format,
			a.body_uri, a.target_source, a.target_hash, a.target_title,
			a.selector_json, a.tags_json, a.created_at, a.indexed_at, a.cid
		FROM all_annotations a
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM likes WHERE subject_uri = a.uri
		) l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM replies WHERE root_uri = a.uri
		) r ON true
		WHERE a.created_at > $1 AND (l.cnt + r.cnt) > 0
		ORDER BY (l.cnt + r.cnt) DESC, a.created_at DESC
		LIMIT $2 OFFSET $3
	`, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetShelvedAnnotations(limit, offset int) ([]Annotation, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			a.uri, a.author_did, a.motivation, a.body_value, a.body_format,
			a.body_uri, a.target_source, a.target_hash, a.target_title,
			a.selector_json, a.tags_json, a.created_at, a.indexed_at, a.cid
		FROM all_annotations a
		WHERE a.created_at < $1 AND a.created_at > $2
			AND NOT EXISTS (SELECT 1 FROM likes WHERE subject_uri = a.uri)
			AND NOT EXISTS (SELECT 1 FROM replies WHERE root_uri = a.uri)
		ORDER BY RANDOM()
		LIMIT $3 OFFSET $4
	`, olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotations(limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByTag(tag string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $1) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) DeleteAnnotation(uri string) error {
	_, err := db.Exec(`DELETE FROM annotations WHERE uri = $1`, uri)
	return err
}

func (db *DB) UpdateAnnotation(uri, bodyValue, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE annotations
		SET body_value = $1, tags_json = $2, cid = $3, indexed_at = $4
		WHERE uri = $5
	`, bodyValue, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) GetAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetMarginAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri NOT LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetSembleAnnotationsByTagAndAuthor(tag, authorDID string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND EXISTS(SELECT 1 FROM jsonb_array_elements_text(tags_json::jsonb) elem WHERE lower(elem) = $2) AND uri LIKE '%network.cosmik%'
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationsByAuthorAndTargetHash(authorDID, targetHash string, limit, offset int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE author_did = $1 AND target_hash = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, authorDID, targetHash, limit, offset)
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

	query := `
		SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM all_annotations
		WHERE uri = ANY($1)
	`

	rows, err := db.Query(query, pqStringArray(uris))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnotations(rows)
}

func (db *DB) GetAnnotationURIs(authorDID string) ([]string, error) {
	rows, err := db.Query(`
		SELECT uri FROM all_annotations WHERE author_did = $1
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
