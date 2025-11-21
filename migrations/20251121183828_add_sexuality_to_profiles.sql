-- +goose Up
-- Add sexuality support to user profiles

-- 1. Create sexualities lookup table
CREATE TABLE IF NOT EXISTS sexualities (
    id SMALLSERIAL PRIMARY KEY,
    label TEXT UNIQUE NOT NULL
);

-- 2. Add sexuality_id column to user_profiles
ALTER TABLE user_profiles
ADD COLUMN IF NOT EXISTS sexuality_id SMALLINT REFERENCES sexualities(id);

-- 3. Seed sexualities table
INSERT INTO sexualities (label) VALUES
    ('Straight'),
    ('Gay'),
    ('Lesbian'),
    ('Bisexual'),
    ('Discovering'),
    ('Not Listed')
ON CONFLICT (label) DO NOTHING;

-- 4. Set default sexuality to "Straight" for existing users
UPDATE user_profiles
SET sexuality_id = (SELECT id FROM sexualities WHERE label = 'Straight' LIMIT 1)
WHERE sexuality_id IS NULL;

-- 5. Create index for sexuality_id
CREATE INDEX IF NOT EXISTS idx_user_profiles_sexuality ON user_profiles(sexuality_id);

-- +goose Down
-- Remove sexuality support

DROP INDEX IF EXISTS idx_user_profiles_sexuality;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS sexuality_id;
DELETE FROM sexualities WHERE label IN ('Straight', 'Gay', 'Lesbian', 'Bisexual', 'Discovering', 'Not Listed');
DROP TABLE IF EXISTS sexualities;

