-- +goose Up
-- Remove dating intentions lookup and all profile/preference references.

DROP INDEX IF EXISTS idx_user_profiles_intention;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS dating_intention_id;

DROP INDEX IF EXISTS idx_prefs_seek_intention;
ALTER TABLE user_preferences DROP COLUMN IF EXISTS seek_intention_ids;

DROP TABLE IF EXISTS dating_intentions;

-- +goose Down

CREATE TABLE IF NOT EXISTS dating_intentions (
    id SMALLSERIAL PRIMARY KEY,
    label TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL DEFAULT ''
);

ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS dating_intention_id SMALLINT REFERENCES dating_intentions (id);

CREATE INDEX IF NOT EXISTS idx_user_profiles_intention ON user_profiles (dating_intention_id);

ALTER TABLE user_preferences
    ADD COLUMN IF NOT EXISTS seek_intention_ids INT[];

CREATE INDEX IF NOT EXISTS idx_prefs_seek_intention ON user_preferences USING GIN (seek_intention_ids);
