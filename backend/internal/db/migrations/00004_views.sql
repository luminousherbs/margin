-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW all_highlights AS
SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
FROM highlights
UNION ALL
SELECT uri, author_did, target_source, target_hash, target_title, selector_json, color, tags_json, created_at, indexed_at, cid
FROM notes WHERE motivation = 'highlighting';
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW all_annotations AS
SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
FROM annotations
UNION ALL
SELECT uri, author_did, motivation, body_value, body_format, body_uri, target_source, target_hash, target_title, selector_json, tags_json, created_at, indexed_at, cid
FROM notes WHERE motivation NOT IN ('highlighting', 'bookmarking');
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE VIEW all_bookmarks AS
SELECT uri, author_did, source, source_hash, title, description, tags_json, created_at, indexed_at, cid
FROM bookmarks
UNION ALL
SELECT uri, author_did, target_source AS source, target_hash AS source_hash, target_title AS title,
    COALESCE(body_value, description) AS description, tags_json, created_at, indexed_at, cid
FROM notes WHERE motivation = 'bookmarking';
-- +goose StatementEnd

-- +goose Down
DROP VIEW IF EXISTS all_bookmarks;
DROP VIEW IF EXISTS all_annotations;
DROP VIEW IF EXISTS all_highlights;
