package db

func (db *DB) CreateReply(r *Reply) error {
	_, err := db.Exec(`
		INSERT INTO replies (uri, author_did, parent_uri, root_uri, text, format, created_at, indexed_at, cid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT(uri) DO UPDATE SET
			text = EXCLUDED.text,
			format = EXCLUDED.format,
			indexed_at = EXCLUDED.indexed_at,
			cid = EXCLUDED.cid
	`, r.URI, r.AuthorDID, r.ParentURI, r.RootURI, r.Text, r.Format, r.CreatedAt, r.IndexedAt, r.CID)
	return err
}

func (db *DB) GetRepliesByRoot(rootURI string) ([]Reply, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, parent_uri, root_uri, text, format, created_at, indexed_at, cid
		FROM replies
		WHERE root_uri = $1
		ORDER BY created_at ASC
	`, rootURI)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReplies(rows)
}

func (db *DB) GetReplyByURI(uri string) (*Reply, error) {
	var r Reply
	err := db.QueryRow(`
		SELECT uri, author_did, parent_uri, root_uri, text, format, created_at, indexed_at, cid
		FROM replies
		WHERE uri = $1
	`, uri).Scan(&r.URI, &r.AuthorDID, &r.ParentURI, &r.RootURI, &r.Text, &r.Format, &r.CreatedAt, &r.IndexedAt, &r.CID)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) DeleteReply(uri string) error {
	_, err := db.Exec(`DELETE FROM replies WHERE uri = $1`, uri)
	return err
}

func (db *DB) GetRepliesByAuthor(authorDID string) ([]Reply, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, parent_uri, root_uri, text, format, created_at, indexed_at, cid
		FROM replies
		WHERE author_did = $1
		ORDER BY created_at DESC
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReplies(rows)
}

func (db *DB) GetOrphanedRepliesByAuthor(authorDID string) ([]Reply, error) {
	rows, err := db.Query(`
		SELECT r.uri, r.author_did, r.parent_uri, r.root_uri, r.text, r.format, r.created_at, r.indexed_at, r.cid
		FROM replies r
		LEFT JOIN annotations a ON r.root_uri = a.uri
		WHERE r.author_did = $1 AND a.uri IS NULL
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReplies(rows)
}

func (db *DB) GetReplyCount(rootURI string) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM replies WHERE root_uri = $1`, rootURI).Scan(&count)
	return count, err
}

func (db *DB) GetReplyCounts(rootURIs []string) (map[string]int, error) {
	if len(rootURIs) == 0 {
		return map[string]int{}, nil
	}

	query := `
		SELECT root_uri, COUNT(*)
		FROM replies
		WHERE root_uri = ANY($1)
		GROUP BY root_uri
	`

	rows, err := db.Query(query, pqStringArray(rootURIs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var uri string
		var count int
		if err := rows.Scan(&uri, &count); err != nil {
			return nil, err
		}
		counts[uri] = count
	}

	return counts, nil
}

func (db *DB) GetRepliesByURIs(uris []string) ([]Reply, error) {
	if len(uris) == 0 {
		return []Reply{}, nil
	}

	rows, err := db.Query(`
		SELECT uri, author_did, parent_uri, root_uri, text, format, created_at, indexed_at, cid
		FROM replies
		WHERE uri = ANY($1)
	`, pqStringArray(uris))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanReplies(rows)
}

func scanReplies(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]Reply, error) {
	var replies []Reply
	for rows.Next() {
		var r Reply
		if err := rows.Scan(&r.URI, &r.AuthorDID, &r.ParentURI, &r.RootURI, &r.Text, &r.Format, &r.CreatedAt, &r.IndexedAt, &r.CID); err != nil {
			return nil, err
		}
		replies = append(replies, r)
	}
	return replies, nil
}
