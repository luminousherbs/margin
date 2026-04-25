-- +goose Up
CREATE TABLE IF NOT EXISTS taken_down_uris (
    uri TEXT PRIMARY KEY,
    taken_down_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS taken_down_uris;
