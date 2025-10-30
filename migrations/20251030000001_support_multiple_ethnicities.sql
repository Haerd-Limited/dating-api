-- +goose Up
-- Create user_ethnicities join table for supporting multiple ethnicities per user
CREATE TABLE user_ethnicities (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    ethnicity_id SMALLINT REFERENCES ethnicities(id),
    PRIMARY KEY (user_id, ethnicity_id)
);
CREATE INDEX idx_user_ethnicities_ethnicity ON user_ethnicities(ethnicity_id);

-- Migrate existing data from user_profiles.ethnicity_id to user_ethnicities
-- Only insert rows where ethnicity_id is not null
INSERT INTO user_ethnicities (user_id, ethnicity_id)
SELECT user_id, ethnicity_id
FROM user_profiles
WHERE ethnicity_id IS NOT NULL;

-- Drop the ethnicity_id column from user_profiles
ALTER TABLE user_profiles DROP COLUMN ethnicity_id;

-- +goose Down
-- Re-add the ethnicity_id column to user_profiles
ALTER TABLE user_profiles ADD COLUMN ethnicity_id SMALLINT REFERENCES ethnicities(id);

-- Migrate data back from user_ethnicities to user_profiles
-- Take the first ethnicity if multiple exist for a user
UPDATE user_profiles up
SET ethnicity_id = (
    SELECT ethnicity_id
    FROM user_ethnicities ue
    WHERE ue.user_id = up.user_id
    ORDER BY ethnicity_id
    LIMIT 1
);

-- Drop the user_ethnicities table
DROP INDEX IF EXISTS idx_user_ethnicities_ethnicity;
DROP TABLE IF EXISTS user_ethnicities;

