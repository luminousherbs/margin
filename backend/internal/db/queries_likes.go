package db

import "fmt"

func (db *DB) CreateLike(l *Like) error {
	_, err := db.Exec(`
		INSERT INTO likes (uri, author_did, subject_uri, created_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(uri) DO NOTHING
	`, l.URI, l.AuthorDID, l.SubjectURI, l.CreatedAt, l.IndexedAt)
	return err
}

func (db *DB) DeleteLike(uri string) error {
	_, err := db.Exec(`DELETE FROM likes WHERE uri = $1`, uri)
	return err
}

func (db *DB) GetLikesByAuthor(authorDID string) ([]Like, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, subject_uri, created_at, indexed_at
		FROM likes
		WHERE author_did = $1
		ORDER BY created_at DESC
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var likes []Like
	for rows.Next() {
		var l Like
		if err := rows.Scan(&l.URI, &l.AuthorDID, &l.SubjectURI, &l.CreatedAt, &l.IndexedAt); err != nil {
			return nil, err
		}
		likes = append(likes, l)
	}
	return likes, nil
}

func (db *DB) GetLikeCount(subjectURI string) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM likes WHERE subject_uri = $1`, subjectURI).Scan(&count)
	return count, err
}

func (db *DB) GetLikeByUserAndSubject(userDID, subjectURI string) (*Like, error) {
	var like Like
	err := db.QueryRow(`
		SELECT uri, author_did, subject_uri, created_at, indexed_at
		FROM likes
		WHERE author_did = $1 AND subject_uri = $2
	`, userDID, subjectURI).Scan(&like.URI, &like.AuthorDID, &like.SubjectURI, &like.CreatedAt, &like.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (db *DB) GetLikeCounts(subjectURIs []string) (map[string]int, error) {
	if len(subjectURIs) == 0 {
		return map[string]int{}, nil
	}

	query := `
		SELECT subject_uri, COUNT(*)
		FROM likes
		WHERE subject_uri IN (` + buildPlaceholders(len(subjectURIs), 1) + `)
		GROUP BY subject_uri
	`

	args := make([]interface{}, len(subjectURIs))
	for i, uri := range subjectURIs {
		args[i] = uri
	}

	rows, err := db.Query(query, args...)
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

func (db *DB) GetViewerLikes(viewerDID string, subjectURIs []string) (map[string]bool, error) {
	if len(subjectURIs) == 0 {
		return map[string]bool{}, nil
	}

	query := fmt.Sprintf(`
		SELECT subject_uri
		FROM likes
		WHERE author_did = $1 AND subject_uri IN (%s)
	`, buildPlaceholders(len(subjectURIs), 2))

	args := make([]interface{}, len(subjectURIs)+1)
	args[0] = viewerDID
	for i, uri := range subjectURIs {
		args[i+1] = uri
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	likes := make(map[string]bool)
	for rows.Next() {
		var uri string
		if err := rows.Scan(&uri); err != nil {
			return nil, err
		}
		likes[uri] = true
	}

	return likes, nil
}
