-- +goose Up
CREATE TABLE IF NOT EXISTS banned_accounts (
    did TEXT PRIMARY KEY,
    reason TEXT,
    banned_by TEXT NOT NULL,
    banned_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS banned_accounts;
