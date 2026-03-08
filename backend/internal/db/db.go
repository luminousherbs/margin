package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
	driver string
}

type Annotation struct {
	URI          string    `json:"uri"`
	AuthorDID    string    `json:"authorDid"`
	Motivation   string    `json:"motivation,omitempty"`
	BodyValue    *string   `json:"bodyValue,omitempty"`
	BodyFormat   *string   `json:"bodyFormat,omitempty"`
	BodyURI      *string   `json:"bodyUri,omitempty"`
	TargetSource string    `json:"targetSource"`
	TargetHash   string    `json:"targetHash"`
	TargetTitle  *string   `json:"targetTitle,omitempty"`
	SelectorJSON *string   `json:"selector,omitempty"`
	TagsJSON     *string   `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	IndexedAt    time.Time `json:"indexedAt"`
	CID          *string   `json:"cid,omitempty"`
}

type Selector struct {
	Type   string `json:"type"`
	Exact  string `json:"exact,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`
	Start  *int   `json:"start,omitempty"`
	End    *int   `json:"end,omitempty"`
	Value  string `json:"value,omitempty"`
}

type Highlight struct {
	URI          string    `json:"uri"`
	AuthorDID    string    `json:"authorDid"`
	TargetSource string    `json:"targetSource"`
	TargetHash   string    `json:"targetHash"`
	TargetTitle  *string   `json:"targetTitle,omitempty"`
	SelectorJSON *string   `json:"selector,omitempty"`
	Color        *string   `json:"color,omitempty"`
	TagsJSON     *string   `json:"tags,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	IndexedAt    time.Time `json:"indexedAt"`
	CID          *string   `json:"cid,omitempty"`
}

type Bookmark struct {
	URI         string    `json:"uri"`
	AuthorDID   string    `json:"authorDid"`
	Source      string    `json:"source"`
	SourceHash  string    `json:"sourceHash"`
	Title       *string   `json:"title,omitempty"`
	Description *string   `json:"description,omitempty"`
	TagsJSON    *string   `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	IndexedAt   time.Time `json:"indexedAt"`
	CID         *string   `json:"cid,omitempty"`
}

type Reply struct {
	URI       string    `json:"uri"`
	AuthorDID string    `json:"authorDid"`
	ParentURI string    `json:"parentUri"`
	RootURI   string    `json:"rootUri"`
	Text      string    `json:"text"`
	Format    *string   `json:"format,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	IndexedAt time.Time `json:"indexedAt"`
	CID       *string   `json:"cid,omitempty"`
}

type Like struct {
	URI        string    `json:"uri"`
	AuthorDID  string    `json:"authorDid"`
	SubjectURI string    `json:"subjectUri"`
	CreatedAt  time.Time `json:"createdAt"`
	IndexedAt  time.Time `json:"indexedAt"`
}

type Collection struct {
	URI         string    `json:"uri"`
	AuthorDID   string    `json:"authorDid"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Icon        *string   `json:"icon,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	IndexedAt   time.Time `json:"indexedAt"`
}

type CollectionItem struct {
	URI           string    `json:"uri"`
	AuthorDID     string    `json:"authorDid"`
	CollectionURI string    `json:"collectionUri"`
	AnnotationURI string    `json:"annotationUri"`
	Position      int       `json:"position"`
	CreatedAt     time.Time `json:"createdAt"`
	IndexedAt     time.Time `json:"indexedAt"`
}

type Notification struct {
	ID           int        `json:"id"`
	RecipientDID string     `json:"recipientDid"`
	ActorDID     string     `json:"actorDid"`
	Type         string     `json:"type"`
	SubjectURI   string     `json:"subjectUri"`
	CreatedAt    time.Time  `json:"createdAt"`
	ReadAt       *time.Time `json:"readAt,omitempty"`
}

type APIKey struct {
	ID         string     `json:"id"`
	OwnerDID   string     `json:"ownerDid"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	URI        string     `json:"uri"`
	CID        *string    `json:"cid,omitempty"`
	IndexedAt  time.Time  `json:"indexedAt"`
}

type Profile struct {
	URI         string    `json:"uri"`
	AuthorDID   string    `json:"authorDid"`
	DisplayName *string   `json:"displayName,omitempty"`
	Avatar      *string   `json:"avatar,omitempty"`
	Bio         *string   `json:"bio,omitempty"`
	Website     *string   `json:"website,omitempty"`
	LinksJSON   *string   `json:"links,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	IndexedAt   time.Time `json:"indexedAt"`
	CID         *string   `json:"cid,omitempty"`
}

type Preferences struct {
	URI                          string    `json:"uri"`
	AuthorDID                    string    `json:"authorDid"`
	ExternalLinkSkippedHostnames *string   `json:"externalLinkSkippedHostnames,omitempty"`
	SubscribedLabelers           *string   `json:"subscribedLabelers,omitempty"`
	LabelPreferences             *string   `json:"labelPreferences,omitempty"`
	DisableExternalLinkWarning   *bool     `json:"disableExternalLinkWarning,omitempty"`
	CreatedAt                    time.Time `json:"createdAt"`
	IndexedAt                    time.Time `json:"indexedAt"`
	CID                          *string   `json:"cid,omitempty"`
}

type Block struct {
	ID         int       `json:"id"`
	ActorDID   string    `json:"actorDid"`
	SubjectDID string    `json:"subjectDid"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Mute struct {
	ID         int       `json:"id"`
	ActorDID   string    `json:"actorDid"`
	SubjectDID string    `json:"subjectDid"`
	CreatedAt  time.Time `json:"createdAt"`
}

type ModerationReport struct {
	ID          int        `json:"id"`
	ReporterDID string     `json:"reporterDid"`
	SubjectDID  string     `json:"subjectDid"`
	SubjectURI  *string    `json:"subjectUri,omitempty"`
	ReasonType  string     `json:"reasonType"`
	ReasonText  *string    `json:"reasonText,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	ResolvedAt  *time.Time `json:"resolvedAt,omitempty"`
	ResolvedBy  *string    `json:"resolvedBy,omitempty"`
}

type ModerationAction struct {
	ID        int       `json:"id"`
	ReportID  int       `json:"reportId"`
	ActorDID  string    `json:"actorDid"`
	Action    string    `json:"action"`
	Comment   *string   `json:"comment,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type ContentLabel struct {
	ID        int       `json:"id"`
	Src       string    `json:"src"`
	URI       string    `json:"uri"`
	Val       string    `json:"val"`
	Neg       bool      `json:"neg"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

func New(dsn string) (*DB, error) {
	driver := "sqlite3"
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		driver = "postgres"
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if driver == "sqlite3" {
		if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
			return nil, fmt.Errorf("failed to set WAL mode: %w", err)
		}
		db.Exec("PRAGMA synchronous=NORMAL;")
		db.Exec("PRAGMA busy_timeout=5000;")
		db.Exec("PRAGMA cache_size=-2000;")
		db.Exec("PRAGMA foreign_keys=ON;")

		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)
	} else {
		db.SetMaxOpenConns(50)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(10 * time.Minute)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db, driver: driver}, nil
}

func (db *DB) Migrate() error {

	dateType := "DATETIME"
	if db.driver == "postgres" {
		dateType = "TIMESTAMP"
	}

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS annotations (
			uri TEXT PRIMARY KEY,
			author_did TEXT NOT NULL,
			motivation TEXT,
			body_value TEXT,
			body_format TEXT DEFAULT 'text/plain',
			body_uri TEXT,
			target_source TEXT NOT NULL,
			target_hash TEXT NOT NULL,
			target_title TEXT,
			selector_json TEXT,
			tags_json TEXT,
			created_at ` + dateType + ` NOT NULL,
			indexed_at ` + dateType + ` NOT NULL,
			cid TEXT
		)`)
	if err != nil {
		return err
	}

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_annotations_target_hash ON annotations(target_hash)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_annotations_target_source ON annotations(target_source)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_annotations_author_did ON annotations(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_annotations_motivation ON annotations(motivation)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_annotations_created_at ON annotations(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS highlights (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		target_source TEXT NOT NULL,
		target_hash TEXT NOT NULL,
		target_title TEXT,
		selector_json TEXT,
		color TEXT,
		tags_json TEXT,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL,
		cid TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_highlights_target_hash ON highlights(target_hash)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_highlights_author_did ON highlights(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_highlights_created_at ON highlights(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS bookmarks (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		source TEXT NOT NULL,
		source_hash TEXT NOT NULL,
		title TEXT,
		description TEXT,
		tags_json TEXT,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL,
		cid TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_source_hash ON bookmarks(source_hash)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_author_did ON bookmarks(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS replies (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		parent_uri TEXT NOT NULL,
		root_uri TEXT NOT NULL,
		text TEXT NOT NULL,
		format TEXT DEFAULT 'text/plain',
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL,
		cid TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_replies_parent_uri ON replies(parent_uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_replies_root_uri ON replies(root_uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_replies_created_at ON replies(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS likes (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		subject_uri TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_likes_subject_uri ON likes(subject_uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_likes_author_did ON likes(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_likes_author_subject ON likes(author_did, subject_uri)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS collections (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		icon TEXT,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_collections_author_did ON collections(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_collections_created_at ON collections(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS collection_items (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		collection_uri TEXT NOT NULL,
		annotation_uri TEXT NOT NULL,
		position INTEGER DEFAULT 0,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_collection_items_collection ON collection_items(collection_uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_collection_items_annotation ON collection_items(annotation_uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_collection_items_created_at ON collection_items(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		did TEXT NOT NULL,
		handle TEXT NOT NULL,
		access_token TEXT NOT NULL,
		refresh_token TEXT NOT NULL,
		dpop_key TEXT,
		created_at ` + dateType + ` NOT NULL,
		expires_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_sessions_did ON sessions(did)`)

	autoInc := "INTEGER PRIMARY KEY AUTOINCREMENT"
	if db.driver == "postgres" {
		autoInc = "SERIAL PRIMARY KEY"
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS edit_history (
		id ` + autoInc + `,
		uri TEXT NOT NULL,
		record_type TEXT NOT NULL,
		previous_content TEXT NOT NULL,
		previous_cid TEXT,
		edited_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_edit_history_uri ON edit_history(uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_edit_history_edited_at ON edit_history(edited_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id ` + autoInc + `,
		recipient_did TEXT NOT NULL,
		actor_did TEXT NOT NULL,
		type TEXT NOT NULL,
		subject_uri TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL,
		read_at ` + dateType + `
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS api_keys (
		id TEXT PRIMARY KEY,
		owner_did TEXT NOT NULL,
		name TEXT NOT NULL,
		key_hash TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL,
		last_used_at ` + dateType + `,
		uri TEXT,
		cid TEXT,
		indexed_at ` + dateType + ` DEFAULT CURRENT_TIMESTAMP
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_api_keys_owner ON api_keys(owner_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS profiles (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		display_name TEXT,
		avatar TEXT,
		bio TEXT,
		website TEXT,
		links_json TEXT,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL,
		cid TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_profiles_author_did ON profiles(author_did)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS preferences (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		external_link_skipped_hostnames TEXT,
		subscribed_labelers TEXT,
		label_preferences TEXT,
		disable_external_link_warning BOOLEAN,
		created_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL,
		cid TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_preferences_author_did ON preferences(author_did)`)

	db.runMigrations()

	db.Exec(`CREATE TABLE IF NOT EXISTS cursors (
		id TEXT PRIMARY KEY,
		last_cursor BIGINT NOT NULL,
		updated_at ` + dateType + ` NOT NULL
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS blocks (
		id ` + autoInc + `,
		actor_did TEXT NOT NULL,
		subject_did TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL,
		UNIQUE(actor_did, subject_did)
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_blocks_actor ON blocks(actor_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_blocks_subject ON blocks(subject_did)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS mutes (
		id ` + autoInc + `,
		actor_did TEXT NOT NULL,
		subject_did TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL,
		UNIQUE(actor_did, subject_did)
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mutes_actor ON mutes(actor_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mutes_subject ON mutes(subject_did)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS moderation_reports (
		id ` + autoInc + `,
		reporter_did TEXT NOT NULL,
		subject_did TEXT NOT NULL,
		subject_uri TEXT,
		reason_type TEXT NOT NULL,
		reason_text TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		created_at ` + dateType + ` NOT NULL,
		resolved_at ` + dateType + `,
		resolved_by TEXT
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_status ON moderation_reports(status)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_subject ON moderation_reports(subject_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_reporter ON moderation_reports(reporter_did)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS moderation_actions (
		id ` + autoInc + `,
		report_id INTEGER NOT NULL,
		actor_did TEXT NOT NULL,
		action TEXT NOT NULL,
		comment TEXT,
		created_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_actions_report ON moderation_actions(report_id)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS content_labels (
		id ` + autoInc + `,
		src TEXT NOT NULL,
		uri TEXT NOT NULL,
		val TEXT NOT NULL,
		neg INTEGER NOT NULL DEFAULT 0,
		created_by TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_content_labels_uri ON content_labels(uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_content_labels_src ON content_labels(src)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS publications (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		url TEXT NOT NULL,
		name TEXT NOT NULL,
		description TEXT,
		show_in_discover BOOLEAN NOT NULL DEFAULT true,
		indexed_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_publications_author ON publications(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_publications_url ON publications(url)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS documents (
		uri TEXT PRIMARY KEY,
		author_did TEXT NOT NULL,
		site TEXT NOT NULL,
		path TEXT,
		title TEXT NOT NULL,
		description TEXT,
		text_content TEXT,
		tags_json TEXT,
		canonical_url TEXT,
		published_at ` + dateType + ` NOT NULL,
		indexed_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_author ON documents(author_did)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_site ON documents(site)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_canonical ON documents(canonical_url)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_documents_published ON documents(published_at DESC)`)

	db.runMigrations()

	return nil
}

func (db *DB) GetProfilesByDIDs(dids []string) (map[string]*Profile, error) {
	if len(dids) == 0 {
		return nil, nil
	}

	query := `SELECT uri, author_did, display_name, bio, avatar, website, links_json, created_at, indexed_at FROM profiles WHERE author_did IN (`
	args := make([]interface{}, len(dids))
	placeholders := make([]string, len(dids))

	for i, did := range dids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = did
	}

	query += strings.Join(placeholders, ",") + ")"

	if db.driver == "sqlite3" {
		query = strings.ReplaceAll(query, "$", "?")

		placeholders = make([]string, len(dids))
		for i := range dids {
			placeholders[i] = "?"
		}
		query = `SELECT uri, author_did, display_name, bio, avatar, website, links_json, created_at, indexed_at FROM profiles WHERE author_did IN (` + strings.Join(placeholders, ",") + ")"
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	profiles := make(map[string]*Profile)
	for rows.Next() {
		var p Profile
		if err := rows.Scan(&p.URI, &p.AuthorDID, &p.DisplayName, &p.Bio, &p.Avatar, &p.Website, &p.LinksJSON, &p.CreatedAt, &p.IndexedAt); err != nil {
			continue
		}
		profiles[p.AuthorDID] = &p
	}

	return profiles, nil
}

func (db *DB) GetCursor(id string) (int64, error) {
	var cursor int64
	err := db.QueryRow("SELECT last_cursor FROM cursors WHERE id = $1", id).Scan(&cursor)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return cursor, nil
}

func (db *DB) SetCursor(id string, cursor int64) error {
	query := `
		INSERT INTO cursors (id, last_cursor, updated_at) 
		VALUES ($1, $2, $3) 
		ON CONFLICT(id) DO UPDATE SET 
			last_cursor = EXCLUDED.last_cursor, 
			updated_at = EXCLUDED.updated_at
	`
	_, err := db.Exec(query, id, cursor, time.Now())
	return err
}

func (db *DB) GetProfile(did string) (*Profile, error) {
	var p Profile
	err := db.QueryRow("SELECT uri, author_did, display_name, avatar, bio, website, links_json, created_at, indexed_at FROM profiles WHERE author_did = $1", did).Scan(
		&p.URI, &p.AuthorDID, &p.DisplayName, &p.Avatar, &p.Bio, &p.Website, &p.LinksJSON, &p.CreatedAt, &p.IndexedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpsertProfile(p *Profile) error {
	query := `
		INSERT INTO profiles (uri, author_did, display_name, avatar, bio, website, links_json, created_at, indexed_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT(uri) DO UPDATE SET 
			display_name = EXCLUDED.display_name,
			avatar = EXCLUDED.avatar,
			bio = EXCLUDED.bio, 
			website = EXCLUDED.website,
			links_json = EXCLUDED.links_json,
			indexed_at = EXCLUDED.indexed_at
	`
	_, err := db.Exec(db.Rebind(query), p.URI, p.AuthorDID, p.DisplayName, p.Avatar, p.Bio, p.Website, p.LinksJSON, p.CreatedAt, p.IndexedAt)
	return err
}

func (db *DB) DeleteProfile(uri string) error {
	_, err := db.Exec("DELETE FROM profiles WHERE uri = $1", uri)
	return err
}

func (db *DB) DeleteAPIKey(id, ownerDID string) (string, error) {
	var uri string
	err := db.QueryRow("SELECT uri FROM api_keys WHERE id = $1 AND owner_did = $2", id, ownerDID).Scan(&uri)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	_, err = db.Exec("DELETE FROM api_keys WHERE id = $1 AND owner_did = $2", id, ownerDID)
	return uri, err
}

func (db *DB) GetPreferences(did string) (*Preferences, error) {
	var p Preferences
	err := db.QueryRow("SELECT uri, author_did, external_link_skipped_hostnames, subscribed_labelers, label_preferences, disable_external_link_warning, created_at, indexed_at, cid FROM preferences WHERE author_did = $1", did).Scan(
		&p.URI, &p.AuthorDID, &p.ExternalLinkSkippedHostnames, &p.SubscribedLabelers, &p.LabelPreferences, &p.DisableExternalLinkWarning, &p.CreatedAt, &p.IndexedAt, &p.CID,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (db *DB) UpsertPreferences(p *Preferences) error {
	query := `
		INSERT INTO preferences (uri, author_did, external_link_skipped_hostnames, subscribed_labelers, label_preferences, disable_external_link_warning, created_at, indexed_at, cid) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) 
		ON CONFLICT(uri) DO UPDATE SET 
			external_link_skipped_hostnames = EXCLUDED.external_link_skipped_hostnames,
			subscribed_labelers = EXCLUDED.subscribed_labelers,
			label_preferences = EXCLUDED.label_preferences,
			disable_external_link_warning = EXCLUDED.disable_external_link_warning,
			indexed_at = EXCLUDED.indexed_at,
			cid = EXCLUDED.cid
	`
	_, err := db.Exec(db.Rebind(query), p.URI, p.AuthorDID, p.ExternalLinkSkippedHostnames, p.SubscribedLabelers, p.LabelPreferences, p.DisableExternalLinkWarning, p.CreatedAt, p.IndexedAt, p.CID)
	return err
}

func (db *DB) DeleteAPIKeyByURI(uri string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE uri = $1", uri)
	return err
}

func (db *DB) DeletePreferences(uri string) error {
	_, err := db.Exec("DELETE FROM preferences WHERE uri = $1", uri)
	return err
}

func (db *DB) GetAPIKeyURIs(ownerDID string) ([]string, error) {
	rows, err := db.Query(db.Rebind("SELECT uri FROM api_keys WHERE owner_did = ? AND uri IS NOT NULL AND uri != ''"), ownerDID)
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

func (db *DB) GetPreferenceURIs(did string) ([]string, error) {
	rows, err := db.Query(db.Rebind("SELECT uri FROM preferences WHERE author_did = ? AND uri IS NOT NULL AND uri != ''"), did)
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

func (db *DB) runMigrations() {
	dateType := "DATETIME"
	if db.driver == "postgres" {
		dateType = "TIMESTAMP"
	}
	db.Exec(`ALTER TABLE sessions ADD COLUMN dpop_key TEXT`)

	db.Exec(`ALTER TABLE annotations ADD COLUMN motivation TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN body_value TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN body_format TEXT DEFAULT 'text/plain'`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN body_uri TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN target_source TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN target_hash TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN target_title TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN selector_json TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN tags_json TEXT`)
	db.Exec(`ALTER TABLE annotations ADD COLUMN cid TEXT`)

	db.Exec(`UPDATE annotations SET target_source = url WHERE target_source IS NULL AND url IS NOT NULL`)
	db.Exec(`UPDATE annotations SET target_hash = url_hash WHERE target_hash IS NULL AND url_hash IS NOT NULL`)
	db.Exec(`UPDATE annotations SET body_value = text WHERE body_value IS NULL AND text IS NOT NULL`)
	db.Exec(`UPDATE annotations SET target_title = title WHERE target_title IS NULL AND title IS NOT NULL`)
	db.Exec(`UPDATE annotations SET motivation = 'commenting' WHERE motivation IS NULL`)

	db.Exec(`ALTER TABLE profiles ADD COLUMN website TEXT`)
	db.Exec(`ALTER TABLE profiles ADD COLUMN display_name TEXT`)
	db.Exec(`ALTER TABLE profiles ADD COLUMN avatar TEXT`)

	if db.driver == "postgres" {
		db.Exec(`ALTER TABLE cursors ALTER COLUMN last_cursor TYPE BIGINT`)
	}

	db.Exec(`ALTER TABLE api_keys ADD COLUMN uri TEXT`)
	db.Exec(`ALTER TABLE api_keys ADD COLUMN cid TEXT`)
	db.Exec(`ALTER TABLE api_keys ADD COLUMN indexed_at ` + dateType + ` DEFAULT CURRENT_TIMESTAMP`)

	db.migrateModeration(dateType)

	db.Exec(`ALTER TABLE preferences ADD COLUMN subscribed_labelers TEXT`)
	db.Exec(`ALTER TABLE preferences ADD COLUMN label_preferences TEXT`)
	db.Exec(`ALTER TABLE preferences ADD COLUMN disable_external_link_warning BOOLEAN`)
}

func (db *DB) migrateModeration(dateType string) {
	_, err := db.Exec(`SELECT subject_did FROM moderation_reports LIMIT 0`)
	if err != nil {
		db.Exec(`DROP TABLE IF EXISTS moderation_reports`)
		db.Exec(`DROP TABLE IF EXISTS moderation_actions`)

		autoInc := "INTEGER PRIMARY KEY AUTOINCREMENT"
		if db.driver == "postgres" {
			autoInc = "SERIAL PRIMARY KEY"
		}

		db.Exec(`CREATE TABLE IF NOT EXISTS moderation_reports (
			id ` + autoInc + `,
			reporter_did TEXT NOT NULL,
			subject_did TEXT NOT NULL,
			subject_uri TEXT,
			reason_type TEXT NOT NULL,
			reason_text TEXT,
			status TEXT NOT NULL DEFAULT 'pending',
			created_at ` + dateType + ` NOT NULL,
			resolved_at ` + dateType + `,
			resolved_by TEXT
		)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_status ON moderation_reports(status)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_subject ON moderation_reports(subject_did)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_reports_reporter ON moderation_reports(reporter_did)`)

		db.Exec(`CREATE TABLE IF NOT EXISTS moderation_actions (
			id ` + autoInc + `,
			report_id INTEGER NOT NULL,
			actor_did TEXT NOT NULL,
			action TEXT NOT NULL,
			comment TEXT,
			created_at ` + dateType + ` NOT NULL
		)`)
		db.Exec(`CREATE INDEX IF NOT EXISTS idx_mod_actions_report ON moderation_actions(report_id)`)
	}

	autoInc := "INTEGER PRIMARY KEY AUTOINCREMENT"
	if db.driver == "postgres" {
		autoInc = "SERIAL PRIMARY KEY"
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS content_labels (
		id ` + autoInc + `,
		src TEXT NOT NULL,
		uri TEXT NOT NULL,
		val TEXT NOT NULL,
		neg INTEGER NOT NULL DEFAULT 0,
		created_by TEXT NOT NULL,
		created_at ` + dateType + ` NOT NULL
	)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_content_labels_uri ON content_labels(uri)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_content_labels_src ON content_labels(src)`)
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) Rebind(query string) string {
	if db.driver != "postgres" {
		return query
	}

	if !strings.Contains(query, "?") {
		return query
	}

	var builder strings.Builder
	builder.Grow(len(query) + 20)

	paramCount := 1
	for _, r := range query {
		if r == '?' {
			fmt.Fprintf(&builder, "$%d", paramCount)
			paramCount++
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func ParseSelector(selectorJSON *string) (*Selector, error) {
	if selectorJSON == nil || *selectorJSON == "" {
		return nil, nil
	}
	var s Selector
	err := json.Unmarshal([]byte(*selectorJSON), &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func ParseTags(tagsJSON *string) ([]string, error) {
	if tagsJSON == nil || *tagsJSON == "" {
		return nil, nil
	}
	var tags []string
	err := json.Unmarshal([]byte(*tagsJSON), &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}
