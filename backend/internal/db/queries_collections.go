package db

import "time"

func (db *DB) CreateCollection(c *Collection) error {
	_, err := db.Exec(`
		INSERT INTO collections (uri, author_did, name, description, icon, created_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(uri) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			icon = EXCLUDED.icon,
			indexed_at = EXCLUDED.indexed_at
	`, c.URI, c.AuthorDID, c.Name, c.Description, c.Icon, c.CreatedAt, c.IndexedAt)
	return err
}

func (db *DB) GetCollectionsByAuthor(authorDID string) ([]Collection, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
		WHERE author_did = $1
		ORDER BY created_at DESC
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.URI, &c.AuthorDID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt, &c.IndexedAt); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}
	return collections, nil
}

func (db *DB) GetCollectionByURI(uri string) (*Collection, error) {
	var c Collection
	err := db.QueryRow(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
		WHERE uri = $1
	`, uri).Scan(&c.URI, &c.AuthorDID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt, &c.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) DeleteCollection(uri string) error {
	db.Exec(`DELETE FROM collection_items WHERE collection_uri = $1`, uri)
	_, err := db.Exec(`DELETE FROM collections WHERE uri = $1`, uri)
	return err
}

func (db *DB) AddToCollection(item *CollectionItem) error {
	_, err := db.Exec(`
		INSERT INTO collection_items (uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(uri) DO UPDATE SET
			position = EXCLUDED.position,
			indexed_at = EXCLUDED.indexed_at
	`, item.URI, item.AuthorDID, item.CollectionURI, item.AnnotationURI, item.Position, item.CreatedAt, item.IndexedAt)
	return err
}

func (db *DB) GetCollectionItems(collectionURI string) ([]CollectionItem, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		WHERE collection_uri = $1
		ORDER BY position ASC, created_at DESC
	`, collectionURI)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CollectionItem
	for rows.Next() {
		var item CollectionItem
		if err := rows.Scan(&item.URI, &item.AuthorDID, &item.CollectionURI, &item.AnnotationURI, &item.Position, &item.CreatedAt, &item.IndexedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (db *DB) RemoveFromCollection(uri string) error {
	_, err := db.Exec(`DELETE FROM collection_items WHERE uri = $1`, uri)
	return err
}

func (db *DB) GetRecentCollectionItems(limit, offset int) ([]CollectionItem, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCollectionItems(rows)
}

func (db *DB) GetPopularCollectionItems(limit, offset int) ([]CollectionItem, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			c.uri, c.author_did, c.collection_uri, c.annotation_uri,
			c.position, c.created_at, c.indexed_at
		FROM collection_items c
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM likes WHERE subject_uri = c.annotation_uri
		) l ON true
		LEFT JOIN LATERAL (
			SELECT COUNT(*) as cnt FROM replies WHERE root_uri = c.annotation_uri
		) r ON true
		WHERE c.created_at > $1 AND (l.cnt + r.cnt) > 0
		ORDER BY (l.cnt + r.cnt) DESC, c.created_at DESC
		LIMIT $2 OFFSET $3
	`, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCollectionItems(rows)
}

func (db *DB) GetShelvedCollectionItems(limit, offset int) ([]CollectionItem, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(`
		SELECT
			c.uri, c.author_did, c.collection_uri, c.annotation_uri,
			c.position, c.created_at, c.indexed_at
		FROM collection_items c
		WHERE c.created_at < $1 AND c.created_at > $2
			AND NOT EXISTS (SELECT 1 FROM likes WHERE subject_uri = c.annotation_uri)
			AND NOT EXISTS (SELECT 1 FROM replies WHERE root_uri = c.annotation_uri)
		ORDER BY RANDOM()
		LIMIT $3 OFFSET $4
	`, olderThan, since, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCollectionItems(rows)
}

func (db *DB) GetCollectionItemsByAuthor(authorDID string) ([]CollectionItem, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		WHERE author_did = $1
		ORDER BY created_at DESC
	`, authorDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCollectionItems(rows)
}

func (db *DB) GetCollectionURIsForAnnotation(annotationURI string) ([]string, error) {
	rows, err := db.Query(`
		SELECT collection_uri FROM collection_items WHERE annotation_uri = $1
	`, annotationURI)
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

func (db *DB) GetCollectionItemCounts(uris []string) (map[string]int, error) {
	if len(uris) == 0 {
		return map[string]int{}, nil
	}

	rows, err := db.Query(`
		SELECT collection_uri, COUNT(*)
		FROM collection_items
		WHERE collection_uri = ANY($1)
		GROUP BY collection_uri
	`, pqStringArray(uris))
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

func (db *DB) GetCollectionsByURIs(uris []string) ([]Collection, error) {
	if len(uris) == 0 {
		return []Collection{}, nil
	}

	rows, err := db.Query(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
		WHERE uri = ANY($1)
	`, pqStringArray(uris))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collections []Collection
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.URI, &c.AuthorDID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt, &c.IndexedAt); err != nil {
			return nil, err
		}
		collections = append(collections, c)
	}
	return collections, nil
}

func scanCollectionItems(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]CollectionItem, error) {
	var items []CollectionItem
	for rows.Next() {
		var item CollectionItem
		if err := rows.Scan(&item.URI, &item.AuthorDID, &item.CollectionURI, &item.AnnotationURI, &item.Position, &item.CreatedAt, &item.IndexedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
