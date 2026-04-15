-- +goose Up
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS dpop_key TEXT;

ALTER TABLE annotations ADD COLUMN IF NOT EXISTS motivation TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS body_value TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS body_format TEXT DEFAULT 'text/plain';
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS body_uri TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS target_source TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS target_hash TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS target_title TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS selector_json TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS tags_json TEXT;
ALTER TABLE annotations ADD COLUMN IF NOT EXISTS cid TEXT;

-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'annotations' AND column_name = 'url') THEN
    UPDATE annotations SET target_source = url      WHERE target_source IS NULL AND url      IS NOT NULL;
    UPDATE annotations SET target_hash   = url_hash WHERE target_hash   IS NULL AND url_hash IS NOT NULL;
    UPDATE annotations SET body_value    = text     WHERE body_value    IS NULL AND text     IS NOT NULL;
    UPDATE annotations SET target_title  = title    WHERE target_title  IS NULL AND title    IS NOT NULL;
  END IF;
  UPDATE annotations SET motivation = 'commenting' WHERE motivation IS NULL;
END $$;
-- +goose StatementEnd

ALTER TABLE profiles ADD COLUMN IF NOT EXISTS website TEXT;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS display_name TEXT;
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS avatar TEXT;

ALTER TABLE cursors ALTER COLUMN last_cursor TYPE BIGINT;

ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS uri TEXT;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS cid TEXT;
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE preferences ADD COLUMN IF NOT EXISTS subscribed_labelers TEXT;
ALTER TABLE preferences ADD COLUMN IF NOT EXISTS label_preferences TEXT;
ALTER TABLE preferences ADD COLUMN IF NOT EXISTS disable_external_link_warning BOOLEAN;
ALTER TABLE preferences ADD COLUMN IF NOT EXISTS enable_community_bookmarks BOOLEAN DEFAULT false;

-- +goose Down