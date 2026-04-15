-- +goose Up
CREATE TABLE IF NOT EXISTS document_embeddings (
    document_uri TEXT PRIMARY KEY,
    embedding TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS annotation_embeddings (
    annotation_uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    document_uri TEXT,
    embedding TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ann_emb_author ON annotation_embeddings(author_did);
CREATE INDEX IF NOT EXISTS idx_ann_emb_document ON annotation_embeddings(document_uri);

CREATE TABLE IF NOT EXISTS user_profiles (
    author_did TEXT PRIMARY KEY,
    embedding TEXT NOT NULL,
    tag_affinities TEXT DEFAULT '{}',
    annotation_count INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS annotation_embeddings;
DROP TABLE IF EXISTS document_embeddings;
