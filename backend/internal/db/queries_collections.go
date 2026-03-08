package db

import "time"

func (db *DB) CreateCollection(c *Collection) error {
	_, err := db.Exec(db.Rebind(`
		INSERT INTO collections (uri, author_did, name, description, icon, created_at, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			icon = excluded.icon,
			indexed_at = excluded.indexed_at
	`), c.URI, c.AuthorDID, c.Name, c.Description, c.Icon, c.CreatedAt, c.IndexedAt)
	return err
}

func (db *DB) GetCollectionsByAuthor(authorDID string) ([]Collection, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
		WHERE author_did = ?
		ORDER BY created_at DESC
	`), authorDID)
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
	err := db.QueryRow(db.Rebind(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
		WHERE uri = ?
	`), uri).Scan(&c.URI, &c.AuthorDID, &c.Name, &c.Description, &c.Icon, &c.CreatedAt, &c.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (db *DB) DeleteCollection(uri string) error {

	db.Exec(db.Rebind(`DELETE FROM collection_items WHERE collection_uri = ?`), uri)
	_, err := db.Exec(db.Rebind(`DELETE FROM collections WHERE uri = ?`), uri)
	return err
}

func (db *DB) AddToCollection(item *CollectionItem) error {
	_, err := db.Exec(db.Rebind(`
		INSERT INTO collection_items (uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			position = excluded.position,
			indexed_at = excluded.indexed_at
	`), item.URI, item.AuthorDID, item.CollectionURI, item.AnnotationURI, item.Position, item.CreatedAt, item.IndexedAt)
	return err
}

func (db *DB) GetCollectionItems(collectionURI string) ([]CollectionItem, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		WHERE collection_uri = ?
		ORDER BY position ASC, created_at DESC
	`), collectionURI)
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
	_, err := db.Exec(db.Rebind(`DELETE FROM collection_items WHERE uri = ?`), uri)
	return err
}

func (db *DB) GetRecentCollectionItems(limit, offset int) ([]CollectionItem, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`), limit, offset)
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

func (db *DB) GetPopularCollectionItems(limit, offset int) ([]CollectionItem, error) {
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			c.uri, c.author_did, c.collection_uri, c.annotation_uri, 
			c.position, c.created_at, c.indexed_at
		FROM collection_items c
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = c.annotation_uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = c.annotation_uri
		WHERE c.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) > 0
		ORDER BY (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) DESC, c.created_at DESC
		LIMIT ? OFFSET ?
	`), since, limit, offset)
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

func (db *DB) GetShelvedCollectionItems(limit, offset int) ([]CollectionItem, error) {
	olderThan := time.Now().AddDate(0, 0, -1)
	since := time.Now().AddDate(0, 0, -14)
	rows, err := db.Query(db.Rebind(`
		SELECT 
			c.uri, c.author_did, c.collection_uri, c.annotation_uri, 
			c.position, c.created_at, c.indexed_at
		FROM collection_items c
		LEFT JOIN (
			SELECT subject_uri, COUNT(*) as cnt FROM likes GROUP BY subject_uri
		) l ON l.subject_uri = c.annotation_uri
		LEFT JOIN (
			SELECT root_uri, COUNT(*) as cnt FROM replies GROUP BY root_uri
		) r ON r.root_uri = c.annotation_uri
		WHERE c.created_at < ? AND c.created_at > ? AND (COALESCE(l.cnt, 0) + COALESCE(r.cnt, 0)) = 0
		ORDER BY RANDOM()
		LIMIT ? OFFSET ?
	`), olderThan, since, limit, offset)
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

func (db *DB) GetCollectionItemsByAuthor(authorDID string) ([]CollectionItem, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT uri, author_did, collection_uri, annotation_uri, position, created_at, indexed_at
		FROM collection_items
		WHERE author_did = ?
		ORDER BY created_at DESC
	`), authorDID)
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

func (db *DB) GetCollectionURIsForAnnotation(annotationURI string) ([]string, error) {
	rows, err := db.Query(db.Rebind(`
		SELECT collection_uri FROM collection_items WHERE annotation_uri = ?
	`), annotationURI)
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

	query := db.Rebind(`
		SELECT collection_uri, COUNT(*)
		FROM collection_items
		WHERE collection_uri IN (` + buildPlaceholders(len(uris)) + `)
		GROUP BY collection_uri
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

	query := db.Rebind(`
		SELECT uri, author_did, name, description, icon, created_at, indexed_at
		FROM collections
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
