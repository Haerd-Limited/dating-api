-- +goose Up
-- Add description column to dating_intentions table
ALTER TABLE dating_intentions
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

-- Update existing dating intentions to new labels and descriptions
-- Map "Life Partner" -> "Marriage-Minded"
UPDATE dating_intentions
SET label = 'Marriage-Minded',
    description = 'Dating with purpose — hoping to find a life partner and build a future together.'
WHERE label = 'Life Partner';

-- Map "Long-term relationship" -> "Serious Relationship"
UPDATE dating_intentions
SET label = 'Serious Relationship',
    description = 'Ready to build something real with the right person — trust, consistency, and shared growth.'
WHERE label = 'Long-term relationship';

-- Map "Long-term relationship, open to short-term" -> "Meaningful Connection"
UPDATE dating_intentions
SET label = 'Meaningful Connection',
    description = 'Open to meeting someone genuine and seeing where it goes — no pressure, just real chemistry.'
WHERE label = 'Long-term relationship, open to short-term';

-- Map "Figuring out my dating goals" -> "Here for the experiment and vibes"
UPDATE dating_intentions
SET label = 'Here for the experiment and vibes',
    description = 'Exploring the dating experience with an open mind and positive energy.'
WHERE label = 'Figuring out my dating goals';

-- Delete the dating intentions we don't want to keep
-- First, set any user profile references to NULL to avoid constraint violations
UPDATE user_profiles
SET dating_intention_id = NULL
WHERE dating_intention_id IN (
    SELECT id FROM dating_intentions
    WHERE label IN (
        'Short-term relationship, open to long-term',
        'Short-term relationship'
    )
);

DELETE FROM dating_intentions
WHERE label IN (
    'Short-term relationship, open to long-term',
    'Short-term relationship'
);

-- +goose Down
-- Restore original dating intention labels and clear descriptions
UPDATE dating_intentions
SET label = 'Life Partner',
    description = ''
WHERE label = 'Marriage-Minded';

UPDATE dating_intentions
SET label = 'Long-term relationship',
    description = ''
WHERE label = 'Serious Relationship';

UPDATE dating_intentions
SET label = 'Long-term relationship, open to short-term',
    description = ''
WHERE label = 'Meaningful Connection';

UPDATE dating_intentions
SET label = 'Figuring out my dating goals',
    description = ''
WHERE label = 'Here for the experiment and vibes';

-- Restore deleted dating intentions
INSERT INTO dating_intentions (label, description) VALUES
    ('Short-term relationship, open to long-term', ''),
    ('Short-term relationship', '')
    ON CONFLICT (label) DO NOTHING;

-- Remove description column (only if it exists and no data depends on it)
ALTER TABLE dating_intentions
    DROP COLUMN IF EXISTS description;

