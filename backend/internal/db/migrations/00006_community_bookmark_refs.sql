-- +goose Up
CREATE TABLE IF NOT EXISTS community_bookmark_refs (
    note_uri      TEXT PRIMARY KEY,
    community_uri TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS community_bookmark_refs;
