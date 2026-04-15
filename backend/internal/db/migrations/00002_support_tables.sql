-- +goose Up
CREATE TABLE IF NOT EXISTS cursors (
    id TEXT PRIMARY KEY,
    last_cursor BIGINT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS collections (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    icon TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_collections_author_did ON collections(author_did);
CREATE INDEX IF NOT EXISTS idx_collections_created_at ON collections(created_at DESC);

CREATE TABLE IF NOT EXISTS collection_items (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    collection_uri TEXT NOT NULL,
    annotation_uri TEXT NOT NULL,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_collection_items_collection ON collection_items(collection_uri);
CREATE INDEX IF NOT EXISTS idx_collection_items_annotation ON collection_items(annotation_uri);
CREATE INDEX IF NOT EXISTS idx_collection_items_created_at ON collection_items(created_at DESC);

CREATE TABLE IF NOT EXISTS blocks (
    id SERIAL PRIMARY KEY,
    actor_did TEXT NOT NULL,
    subject_did TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(actor_did, subject_did)
);
CREATE INDEX IF NOT EXISTS idx_blocks_actor ON blocks(actor_did);
CREATE INDEX IF NOT EXISTS idx_blocks_subject ON blocks(subject_did);

CREATE TABLE IF NOT EXISTS mutes (
    id SERIAL PRIMARY KEY,
    actor_did TEXT NOT NULL,
    subject_did TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(actor_did, subject_did)
);
CREATE INDEX IF NOT EXISTS idx_mutes_actor ON mutes(actor_did);
CREATE INDEX IF NOT EXISTS idx_mutes_subject ON mutes(subject_did);

CREATE TABLE IF NOT EXISTS content_labels (
    id SERIAL PRIMARY KEY,
    src TEXT NOT NULL,
    uri TEXT NOT NULL,
    val TEXT NOT NULL,
    neg INTEGER NOT NULL DEFAULT 0,
    created_by TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_content_labels_uri ON content_labels(uri);
CREATE INDEX IF NOT EXISTS idx_content_labels_src ON content_labels(src);

CREATE TABLE IF NOT EXISTS moderation_reports (
    id SERIAL PRIMARY KEY,
    reporter_did TEXT NOT NULL,
    subject_did TEXT NOT NULL,
    subject_uri TEXT,
    reason_type TEXT NOT NULL,
    reason_text TEXT,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    resolved_by TEXT
);
CREATE INDEX IF NOT EXISTS idx_mod_reports_status ON moderation_reports(status);
CREATE INDEX IF NOT EXISTS idx_mod_reports_subject ON moderation_reports(subject_did);
CREATE INDEX IF NOT EXISTS idx_mod_reports_reporter ON moderation_reports(reporter_did);

CREATE TABLE IF NOT EXISTS moderation_actions (
    id SERIAL PRIMARY KEY,
    report_id INTEGER NOT NULL,
    actor_did TEXT NOT NULL,
    action TEXT NOT NULL,
    comment TEXT,
    created_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_mod_actions_report ON moderation_actions(report_id);

CREATE TABLE IF NOT EXISTS publications (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    url TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    show_in_discover BOOLEAN NOT NULL DEFAULT true,
    indexed_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_publications_author ON publications(author_did);
CREATE INDEX IF NOT EXISTS idx_publications_url ON publications(url);

CREATE TABLE IF NOT EXISTS documents (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    site TEXT NOT NULL,
    path TEXT,
    title TEXT NOT NULL,
    description TEXT,
    text_content TEXT,
    tags_json TEXT,
    canonical_url TEXT,
    published_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_documents_author ON documents(author_did);
CREATE INDEX IF NOT EXISTS idx_documents_site ON documents(site);
CREATE INDEX IF NOT EXISTS idx_documents_canonical ON documents(canonical_url);
CREATE INDEX IF NOT EXISTS idx_documents_published ON documents(published_at DESC);

-- +goose Down
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS publications;
DROP TABLE IF EXISTS moderation_actions;
DROP TABLE IF EXISTS moderation_reports;
DROP TABLE IF EXISTS content_labels;
DROP TABLE IF EXISTS mutes;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS collection_items;
DROP TABLE IF EXISTS collections;
DROP TABLE IF EXISTS cursors;
