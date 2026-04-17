package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func (db *DB) CreateNote(n *Note) error {
	query := `
		INSERT INTO notes (
			uri, author_did, motivation, color, description, body_value, body_format, body_uri,
			target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		) ON CONFLICT (uri) DO UPDATE SET
			motivation = EXCLUDED.motivation,
			color = EXCLUDED.color,
			description = EXCLUDED.description,
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
	`
	_, err := db.Exec(query,
		n.URI, n.AuthorDID, n.Motivation, n.Color, n.Description, n.BodyValue, n.BodyFormat, n.BodyURI,
		n.TargetSource, n.TargetHash, n.TargetTitle, n.SelectorJSON, n.TagsJSON, n.CreatedAt, n.IndexedAt, n.CID,
	)
	return err
}

func (db *DB) GetNoteByURI(uri string) (*Note, error) {
	query := `
		SELECT uri, author_did, motivation, color, description, body_value, body_format, body_uri,
			target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM notes WHERE uri = $1
	`
	var n Note
	err := db.QueryRow(query, uri).Scan(
		&n.URI, &n.AuthorDID, &n.Motivation, &n.Color, &n.Description, &n.BodyValue, &n.BodyFormat, &n.BodyURI,
		&n.TargetSource, &n.TargetHash, &n.TargetTitle, &n.SelectorJSON, &n.TagsJSON, &n.CreatedAt, &n.IndexedAt, &n.CID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (db *DB) MarginNoteBookmarkExists(authorDID, targetHash string) (bool, error) {
	var dummy int
	err := db.QueryRow(`
		SELECT 1 FROM notes
		WHERE author_did = $1
		  AND target_hash = $2
		  AND motivation = 'bookmarking'
		  AND uri LIKE 'at://%/at.margin.note/%'
		LIMIT 1
	`, authorDID, targetHash).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *DB) GetCommunityBookmarkURI(authorDID, targetHash string) (string, error) {
	var uri string
	err := db.QueryRow(`
		SELECT uri FROM notes
		WHERE author_did = $1
		  AND target_hash = $2
		  AND uri LIKE 'at://%/community.lexicon.bookmarks.bookmark/%'
		LIMIT 1
	`, authorDID, targetHash).Scan(&uri)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return uri, err
}

func (db *DB) CommunityBookmarkExists(authorDID, targetHash, tagsJSON string) (bool, error) {
	query := `
		SELECT 1 FROM notes
		WHERE author_did = $1
		  AND target_hash = $2
		  AND uri LIKE 'at://%/community.lexicon.bookmarks.bookmark/%'
		  AND COALESCE(tags_json, '[]') = COALESCE($3, '[]')
		LIMIT 1
	`
	normalized := tagsJSON
	if normalized == "" {
		normalized = "[]"
	}
	var dummy int
	err := db.QueryRow(query, authorDID, targetHash, normalized).Scan(&dummy)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *DB) GetNotesByURIs(uris []string) ([]Note, error) {
	if len(uris) == 0 {
		return nil, nil
	}
	placeholders := make([]string, len(uris))
	args := make([]interface{}, len(uris))
	for i, u := range uris {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = u
	}
	query := `
		SELECT uri, author_did, motivation, color, description, body_value, body_format, body_uri,
			target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
		FROM notes WHERE uri IN (` + strings.Join(placeholders, ",") + `)`
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []Note
	for rows.Next() {
		var n Note
		if err := rows.Scan(
			&n.URI, &n.AuthorDID, &n.Motivation, &n.Color, &n.Description, &n.BodyValue, &n.BodyFormat, &n.BodyURI,
			&n.TargetSource, &n.TargetHash, &n.TargetTitle, &n.SelectorJSON, &n.TagsJSON, &n.CreatedAt, &n.IndexedAt, &n.CID,
		); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, nil
}

func (db *DB) DeleteNote(uri string) error {
	_, err := db.Exec("DELETE FROM notes WHERE uri = $1", uri)
	return err
}

func (db *DB) UpdateNoteAnnotation(uri, bodyValue, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE notes
		SET body_value = $1, tags_json = NULLIF($2, ''), cid = $3, indexed_at = $4
		WHERE uri = $5
	`, bodyValue, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) UpdateNoteHighlight(uri, color, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE notes
		SET color = NULLIF($1, ''), tags_json = NULLIF($2, ''), cid = $3, indexed_at = $4
		WHERE uri = $5
	`, color, tagsJSON, cid, time.Now(), uri)
	return err
}

func (db *DB) UpdateNoteBookmark(uri, title, description, tagsJSON, cid string) error {
	_, err := db.Exec(`
		UPDATE notes
		SET target_title = NULLIF($1, ''), body_value = NULLIF($2, ''), tags_json = NULLIF($3, ''), cid = $4, indexed_at = $5
		WHERE uri = $6
	`, title, description, tagsJSON, cid, time.Now(), uri)
	return err
}
