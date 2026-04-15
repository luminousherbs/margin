-- +goose Up
CREATE TABLE IF NOT EXISTS notes (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    motivation TEXT,
    color TEXT,
    description TEXT,
    body_value TEXT,
    body_format TEXT DEFAULT 'text/plain',
    body_uri TEXT,
    target_source TEXT NOT NULL,
    target_hash TEXT NOT NULL,
    target_title TEXT,
    selector_json TEXT,
    tags_json TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_notes_target_hash ON notes(target_hash);
CREATE INDEX IF NOT EXISTS idx_notes_target_source ON notes(target_source);
CREATE INDEX IF NOT EXISTS idx_notes_author_did ON notes(author_did);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON notes(created_at DESC);

CREATE TABLE IF NOT EXISTS annotations (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    motivation TEXT,
    body_value TEXT,
    body_format TEXT DEFAULT 'text/plain',
    body_uri TEXT,
    target_source TEXT,
    target_hash TEXT,
    target_title TEXT,
    selector_json TEXT,
    tags_json TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_annotations_target_hash ON annotations(target_hash);
CREATE INDEX IF NOT EXISTS idx_annotations_target_source ON annotations(target_source);
CREATE INDEX IF NOT EXISTS idx_annotations_author_did ON annotations(author_did);
CREATE INDEX IF NOT EXISTS idx_annotations_motivation ON annotations(motivation);
CREATE INDEX IF NOT EXISTS idx_annotations_created_at ON annotations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_annotations_author_created ON annotations(author_did, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_annotations_uri_pattern ON annotations(uri text_pattern_ops);

CREATE TABLE IF NOT EXISTS highlights (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    target_source TEXT NOT NULL,
    target_hash TEXT NOT NULL,
    target_title TEXT,
    selector_json TEXT,
    color TEXT,
    tags_json TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_highlights_target_hash ON highlights(target_hash);
CREATE INDEX IF NOT EXISTS idx_highlights_author_did ON highlights(author_did);
CREATE INDEX IF NOT EXISTS idx_highlights_created_at ON highlights(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_highlights_author_created ON highlights(author_did, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_highlights_uri_pattern ON highlights(uri text_pattern_ops);

CREATE TABLE IF NOT EXISTS bookmarks (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    source TEXT NOT NULL,
    source_hash TEXT NOT NULL,
    title TEXT,
    description TEXT,
    tags_json TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_bookmarks_source_hash ON bookmarks(source_hash);
CREATE INDEX IF NOT EXISTS idx_bookmarks_author_did ON bookmarks(author_did);
CREATE INDEX IF NOT EXISTS idx_bookmarks_created_at ON bookmarks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bookmarks_author_created ON bookmarks(author_did, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_bookmarks_uri_pattern ON bookmarks(uri text_pattern_ops);

CREATE TABLE IF NOT EXISTS replies (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    parent_uri TEXT NOT NULL,
    root_uri TEXT NOT NULL,
    text TEXT NOT NULL,
    format TEXT DEFAULT 'text/plain',
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_replies_parent_uri ON replies(parent_uri);
CREATE INDEX IF NOT EXISTS idx_replies_root_uri ON replies(root_uri);
CREATE INDEX IF NOT EXISTS idx_replies_created_at ON replies(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_replies_author_did ON replies(author_did);

CREATE TABLE IF NOT EXISTS likes (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    subject_uri TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_likes_subject_uri ON likes(subject_uri);
CREATE INDEX IF NOT EXISTS idx_likes_author_did ON likes(author_did);
CREATE INDEX IF NOT EXISTS idx_likes_author_subject ON likes(author_did, subject_uri);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    did TEXT NOT NULL,
    handle TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    dpop_key TEXT,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sessions_did ON sessions(did);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS edit_history (
    id SERIAL PRIMARY KEY,
    uri TEXT NOT NULL,
    record_type TEXT NOT NULL,
    previous_content TEXT NOT NULL,
    previous_cid TEXT,
    edited_at TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_edit_history_uri ON edit_history(uri);
CREATE INDEX IF NOT EXISTS idx_edit_history_edited_at ON edit_history(edited_at DESC);

CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    recipient_did TEXT NOT NULL,
    actor_did TEXT NOT NULL,
    type TEXT NOT NULL,
    subject_uri TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    read_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_did);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(recipient_did) WHERE read_at IS NULL;

CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    owner_did TEXT NOT NULL,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_used_at TIMESTAMP,
    uri TEXT,
    cid TEXT,
    indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_api_keys_owner ON api_keys(owner_did);
CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash);

CREATE TABLE IF NOT EXISTS profiles (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    display_name TEXT,
    avatar TEXT,
    bio TEXT,
    website TEXT,
    links_json TEXT,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_profiles_author_did ON profiles(author_did);

CREATE TABLE IF NOT EXISTS preferences (
    uri TEXT PRIMARY KEY,
    author_did TEXT NOT NULL,
    external_link_skipped_hostnames TEXT,
    subscribed_labelers TEXT,
    label_preferences TEXT,
    disable_external_link_warning BOOLEAN,
    enable_community_bookmarks BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL,
    indexed_at TIMESTAMP NOT NULL,
    cid TEXT
);
CREATE INDEX IF NOT EXISTS idx_preferences_author_did ON preferences(author_did);

-- +goose Down
DROP TABLE IF EXISTS preferences;
DROP TABLE IF EXISTS profiles;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS edit_history;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS replies;
DROP TABLE IF EXISTS bookmarks;
DROP TABLE IF EXISTS highlights;
DROP TABLE IF EXISTS annotations;
DROP TABLE IF EXISTS notes;
