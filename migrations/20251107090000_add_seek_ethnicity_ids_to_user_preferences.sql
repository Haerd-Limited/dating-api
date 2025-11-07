-- +goose Up
ALTER TABLE user_preferences
    ADD COLUMN seek_ethnicity_ids INT[];

CREATE INDEX IF NOT EXISTS idx_prefs_seek_ethnicity
    ON user_preferences USING GIN (seek_ethnicity_ids);

-- +goose Down
DROP INDEX IF EXISTS idx_prefs_seek_ethnicity;

ALTER TABLE user_preferences
    DROP COLUMN IF EXISTS seek_ethnicity_ids;

