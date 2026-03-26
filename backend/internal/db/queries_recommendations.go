package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Document struct {
	URI          string    `json:"uri"`
	AuthorDID    string    `json:"authorDid"`
	Site         string    `json:"site"`
	Path         *string   `json:"path,omitempty"`
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	TextContent  *string   `json:"textContent,omitempty"`
	TagsJSON     *string   `json:"tags,omitempty"`
	CanonicalURL string    `json:"canonicalUrl"`
	PublishedAt  time.Time `json:"publishedAt"`
	IndexedAt    time.Time `json:"indexedAt"`
}

type Publication struct {
	URI            string    `json:"uri"`
	AuthorDID      string    `json:"authorDid"`
	URL            string    `json:"url"`
	Name           string    `json:"name"`
	Description    *string   `json:"description,omitempty"`
	ShowInDiscover bool      `json:"showInDiscover"`
	IndexedAt      time.Time `json:"indexedAt"`
}

type DocumentEmbedding struct {
	DocumentURI string    `json:"documentUri"`
	Embedding   []float32 `json:"embedding"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type AnnotationEmbedding struct {
	AnnotationURI string    `json:"annotationUri"`
	AuthorDID     string    `json:"authorDid"`
	DocumentURI   *string   `json:"documentUri,omitempty"`
	Embedding     []float32 `json:"embedding"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type UserProfile struct {
	AuthorDID       string    `json:"authorDid"`
	Embedding       []float32 `json:"embedding"`
	TagAffinities   string    `json:"tagAffinities"`
	AnnotationCount int       `json:"annotationCount"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (db *DB) MigrateRecommendations() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS document_embeddings (
			document_uri TEXT PRIMARY KEY,
			embedding TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("create document_embeddings table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS annotation_embeddings (
			annotation_uri TEXT PRIMARY KEY,
			author_did TEXT NOT NULL,
			document_uri TEXT,
			embedding TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("create annotation_embeddings table: %w", err)
	}
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_ann_emb_author ON annotation_embeddings(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_ann_emb_document ON annotation_embeddings(document_uri)`)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS user_profiles (
			author_did TEXT PRIMARY KEY,
			embedding TEXT NOT NULL,
			tag_affinities TEXT DEFAULT '{}',
			annotation_count INTEGER NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL
		)`)
	if err != nil {
		return fmt.Errorf("create user_profiles table: %w", err)
	}

	return nil
}

func (db *DB) UpsertPublication(p *Publication) error {
	query := `
		INSERT INTO publications (uri, author_did, url, name, description, show_in_discover, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT(uri) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			show_in_discover = EXCLUDED.show_in_discover,
			indexed_at = EXCLUDED.indexed_at
	`
	_, err := db.Exec(query, p.URI, p.AuthorDID, p.URL, p.Name, p.Description, p.ShowInDiscover, p.IndexedAt)
	return err
}

func (db *DB) DeletePublication(uri string) error {
	_, err := db.Exec("DELETE FROM publications WHERE uri = $1", uri)
	return err
}

func (db *DB) GetPublicationByURL(url string) (*Publication, error) {
	var p Publication
	err := db.QueryRow(
		"SELECT uri, author_did, url, name, description, show_in_discover, indexed_at FROM publications WHERE url = $1",
		url,
	).Scan(&p.URI, &p.AuthorDID, &p.URL, &p.Name, &p.Description, &p.ShowInDiscover, &p.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpsertDocument(d *Document) error {
	query := `
		INSERT INTO documents (uri, author_did, site, path, title, description, text_content, tags_json, canonical_url, published_at, indexed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT(uri) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			text_content = EXCLUDED.text_content,
			tags_json = EXCLUDED.tags_json,
			canonical_url = EXCLUDED.canonical_url,
			indexed_at = EXCLUDED.indexed_at
	`
	_, err := db.Exec(query, d.URI, d.AuthorDID, d.Site, d.Path, d.Title, d.Description, d.TextContent, d.TagsJSON, d.CanonicalURL, d.PublishedAt, d.IndexedAt)
	return err
}

func (db *DB) DeleteDocument(uri string) error {
	_, err := db.Exec("DELETE FROM documents WHERE uri = $1", uri)
	return err
}

func (db *DB) GetDocumentByCanonicalURL(canonicalURL string) (*Document, error) {
	var d Document
	err := db.QueryRow(
		`SELECT uri, author_did, site, path, title, description, text_content, tags_json, canonical_url, published_at, indexed_at
		 FROM documents WHERE canonical_url = $1`,
		canonicalURL,
	).Scan(&d.URI, &d.AuthorDID, &d.Site, &d.Path, &d.Title, &d.Description, &d.TextContent, &d.TagsJSON, &d.CanonicalURL, &d.PublishedAt, &d.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) GetDocumentByURI(uri string) (*Document, error) {
	var d Document
	err := db.QueryRow(
		`SELECT uri, author_did, site, path, title, description, text_content, tags_json, canonical_url, published_at, indexed_at
		 FROM documents WHERE uri = $1`,
		uri,
	).Scan(&d.URI, &d.AuthorDID, &d.Site, &d.Path, &d.Title, &d.Description, &d.TextContent, &d.TagsJSON, &d.CanonicalURL, &d.PublishedAt, &d.IndexedAt)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (db *DB) GetDocumentsWithoutEmbeddings(limit int) ([]Document, error) {
	rows, err := db.Query(`
		SELECT d.uri, d.author_did, d.site, d.path, d.title, d.description, d.text_content, d.tags_json, d.canonical_url, d.published_at, d.indexed_at
		FROM documents d
		LEFT JOIN document_embeddings de ON d.uri = de.document_uri
		WHERE de.document_uri IS NULL
		ORDER BY d.indexed_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func (db *DB) GetAnnotationsWithoutEmbeddings(limit int) ([]Annotation, error) {
	rows, err := db.Query(`
		SELECT a.uri, a.author_did, a.motivation, a.body_value, a.body_format, a.body_uri, a.target_source, a.target_hash, a.target_title, a.selector_json, a.tags_json, a.created_at, a.indexed_at, a.cid
		FROM annotations a
		LEFT JOIN annotation_embeddings ae ON a.uri = ae.annotation_uri
		WHERE ae.annotation_uri IS NULL AND a.motivation IN ('commenting', 'highlighting')
		ORDER BY a.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAnnotations(rows)
}

type HighlightForEmbedding struct {
	URI          string
	AuthorDID    string
	TargetSource string
	TargetTitle  *string
	SelectorJSON *string
	TagsJSON     *string
}

func (db *DB) GetHighlightsWithoutEmbeddings(limit int) ([]HighlightForEmbedding, error) {
	rows, err := db.Query(`
		SELECT h.uri, h.author_did, h.target_source, h.target_title, h.selector_json, h.tags_json
		FROM highlights h
		LEFT JOIN annotation_embeddings ae ON h.uri = ae.annotation_uri
		WHERE ae.annotation_uri IS NULL
		ORDER BY h.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []HighlightForEmbedding
	for rows.Next() {
		var h HighlightForEmbedding
		if err := rows.Scan(&h.URI, &h.AuthorDID, &h.TargetSource, &h.TargetTitle, &h.SelectorJSON, &h.TagsJSON); err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, nil
}

func (db *DB) GetDistinctAnnotationAuthors() ([]string, error) {
	rows, err := db.Query(`SELECT DISTINCT author_did FROM annotation_embeddings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dids []string
	for rows.Next() {
		var did string
		if err := rows.Scan(&did); err != nil {
			return nil, err
		}
		dids = append(dids, did)
	}
	return dids, nil
}

func scanDocuments(rows interface {
	Next() bool
	Scan(...interface{}) error
}) ([]Document, error) {
	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.URI, &d.AuthorDID, &d.Site, &d.Path, &d.Title, &d.Description, &d.TextContent, &d.TagsJSON, &d.CanonicalURL, &d.PublishedAt, &d.IndexedAt); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (db *DB) GetRecentDocuments(limit, offset int) ([]Document, error) {
	rows, err := db.Query(`
		SELECT uri, author_did, site, path, title, description, text_content, tags_json, canonical_url, published_at, indexed_at
		FROM documents
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func (db *DB) GetPopularDocuments(limit, offset int) ([]Document, error) {
	rows, err := db.Query(`
		SELECT d.uri, d.author_did, d.site, d.path, d.title, d.description, d.text_content, d.tags_json, d.canonical_url, d.published_at, d.indexed_at
		FROM documents d
		LEFT JOIN annotations a ON a.target_source = d.canonical_url
		GROUP BY d.uri
		ORDER BY COUNT(a.uri) DESC, d.published_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func (db *DB) GetDocumentCount() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM documents").Scan(&count)
	return count, err
}

func (db *DB) UpsertDocumentEmbedding(documentURI string, embedding []float32) error {
	vecStr := float32SliceToVectorString(embedding)
	_, err := db.Exec(
		`INSERT INTO document_embeddings (document_uri, embedding, updated_at) VALUES ($1, $2, $3)
		 ON CONFLICT(document_uri) DO UPDATE SET embedding = EXCLUDED.embedding, updated_at = EXCLUDED.updated_at`,
		documentURI, vecStr, time.Now(),
	)
	return err
}

func (db *DB) UpsertAnnotationEmbedding(annotationURI, authorDID string, documentURI *string, embedding []float32) error {
	vecStr := float32SliceToVectorString(embedding)
	_, err := db.Exec(
		`INSERT INTO annotation_embeddings (annotation_uri, author_did, document_uri, embedding, updated_at) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT(annotation_uri) DO UPDATE SET embedding = EXCLUDED.embedding, document_uri = EXCLUDED.document_uri, updated_at = EXCLUDED.updated_at`,
		annotationURI, authorDID, documentURI, vecStr, time.Now(),
	)
	return err
}

func (db *DB) DeleteAnnotationEmbedding(annotationURI string) error {
	_, err := db.Exec("DELETE FROM annotation_embeddings WHERE annotation_uri = $1", annotationURI)
	return err
}

func (db *DB) UpsertUserProfile(authorDID string, embedding []float32, tagAffinities map[string]float64, annotationCount int) error {
	vecStr := float32SliceToVectorString(embedding)
	tagsJSON, _ := json.Marshal(tagAffinities)
	_, err := db.Exec(
		`INSERT INTO user_profiles (author_did, embedding, tag_affinities, annotation_count, updated_at) VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT(author_did) DO UPDATE SET embedding = EXCLUDED.embedding, tag_affinities = EXCLUDED.tag_affinities, annotation_count = EXCLUDED.annotation_count, updated_at = EXCLUDED.updated_at`,
		authorDID, vecStr, string(tagsJSON), annotationCount, time.Now(),
	)
	return err
}

func (db *DB) GetUserProfile(authorDID string) (*UserProfile, error) {
	var p UserProfile
	var embStr string
	err := db.QueryRow(
		`SELECT author_did, embedding, tag_affinities, annotation_count, updated_at FROM user_profiles WHERE author_did = $1`,
		authorDID,
	).Scan(&p.AuthorDID, &embStr, &p.TagAffinities, &p.AnnotationCount, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Embedding = parseVectorString(embStr)
	return &p, nil
}

func (db *DB) GetAnnotationEmbeddingsByAuthor(authorDID string) ([]AnnotationEmbedding, error) {
	rows, err := db.Query(
		`SELECT annotation_uri, author_did, document_uri, embedding, updated_at FROM annotation_embeddings WHERE author_did = $1`,
		authorDID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AnnotationEmbedding
	for rows.Next() {
		var ae AnnotationEmbedding
		var embStr string
		if err := rows.Scan(&ae.AnnotationURI, &ae.AuthorDID, &ae.DocumentURI, &embStr, &ae.UpdatedAt); err != nil {
			return nil, err
		}
		ae.Embedding = parseVectorString(embStr)
		results = append(results, ae)
	}
	return results, nil
}

func (db *DB) GetRecentAnnotationEmbeddingsByAuthor(authorDID string, limit int) ([]AnnotationEmbedding, error) {
	rows, err := db.Query(
		`SELECT annotation_uri, author_did, document_uri, embedding, updated_at FROM annotation_embeddings WHERE author_did = $1 ORDER BY updated_at DESC LIMIT $2`,
		authorDID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []AnnotationEmbedding
	for rows.Next() {
		var ae AnnotationEmbedding
		var embStr string
		if err := rows.Scan(&ae.AnnotationURI, &ae.AuthorDID, &ae.DocumentURI, &embStr, &ae.UpdatedAt); err != nil {
			return nil, err
		}
		ae.Embedding = parseVectorString(embStr)
		results = append(results, ae)
	}
	return results, nil
}

type CandidateDocument struct {
	URI          string    `json:"uri"`
	AuthorDID    string    `json:"authorDid"`
	Site         string    `json:"site"`
	Path         *string   `json:"path,omitempty"`
	Title        string    `json:"title"`
	Description  *string   `json:"description,omitempty"`
	TagsJSON     *string   `json:"tags,omitempty"`
	CanonicalURL string    `json:"canonicalUrl"`
	PublishedAt  time.Time `json:"publishedAt"`
	Embedding    []float32 `json:"-"`
	Engagement   int       `json:"engagement"`
}

func (db *DB) GetCandidateDocuments(userDID string, limit int) ([]CandidateDocument, error) {
	rows, err := db.Query(`
		SELECT
			d.uri, d.author_did, d.site, d.path, d.title, d.description, d.tags_json,
			d.canonical_url, d.published_at, de.embedding,
			COALESCE(eng.cnt, 0) AS engagement
		FROM documents d
		JOIN document_embeddings de ON d.uri = de.document_uri
		LEFT JOIN (
			SELECT document_uri, COUNT(DISTINCT author_did) AS cnt
			FROM annotation_embeddings
			WHERE document_uri IS NOT NULL
			GROUP BY document_uri
		) eng ON eng.document_uri = d.uri
		LEFT JOIN publications p ON d.site = p.uri OR d.site = p.url
		WHERE d.author_did != $1
		  AND (p.show_in_discover IS NULL OR p.show_in_discover = true)
		  AND LENGTH(d.title) > 15
		  AND (LENGTH(COALESCE(d.description, '')) >= 30 OR LENGTH(COALESCE(d.text_content, '')) >= 100)
		  AND LOWER(d.title) NOT LIKE '%test%'
		  AND LOWER(d.title) NOT LIKE '%testing%'
		  AND LOWER(d.title) NOT LIKE '%hello world%'
		  AND LOWER(d.title) NOT LIKE '%untitled%'
		  AND LOWER(d.title) NOT LIKE '%draft%'
		  AND LOWER(d.title) NOT LIKE '%asdf%'
		  AND LOWER(d.title) NOT LIKE '%lorem%'
		  AND LOWER(d.title) NOT LIKE '%placeholder%'
		  AND d.uri NOT IN (
		    SELECT DISTINCT document_uri FROM annotation_embeddings
		    WHERE author_did = $2 AND document_uri IS NOT NULL
		  )
		ORDER BY d.published_at DESC
		LIMIT $3
	`, userDID, userDID, limit)
	if err != nil {
		return nil, fmt.Errorf("candidate query: %w", err)
	}
	defer rows.Close()

	var results []CandidateDocument
	for rows.Next() {
		var c CandidateDocument
		var embStr string
		if err := rows.Scan(
			&c.URI, &c.AuthorDID, &c.Site, &c.Path, &c.Title, &c.Description,
			&c.TagsJSON, &c.CanonicalURL, &c.PublishedAt, &embStr, &c.Engagement,
		); err != nil {
			return nil, err
		}
		c.Embedding = parseVectorString(embStr)
		results = append(results, c)
	}
	return results, nil
}

func (db *DB) MatchAnnotationToDocument(targetSource string) (*string, error) {
	var uri string
	err := db.QueryRow(`SELECT uri FROM documents WHERE canonical_url = $1`, targetSource).Scan(&uri)
	if err != nil {
		return nil, err
	}
	return &uri, nil
}

func float32SliceToVectorString(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

func parseVectorString(s string) []float32 {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]float32, len(parts))
	for i, p := range parts {
		f, _ := strconv.ParseFloat(strings.TrimSpace(p), 32)
		result[i] = float32(f)
	}
	return result
}
