-- +goose Up
ALTER TABLE user_preferences
    ADD COLUMN seek_sexuality_ids INT[];

CREATE INDEX IF NOT EXISTS idx_prefs_seek_sexuality
    ON user_preferences USING GIN (seek_sexuality_ids);

-- +goose Down
DROP INDEX IF EXISTS idx_prefs_seek_sexuality;

ALTER TABLE user_preferences
    DROP COLUMN IF EXISTS seek_sexuality_ids;

