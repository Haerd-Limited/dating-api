-- +goose Up
-- Map users with NULL dating_intention_id to "Meaningful Connection"
-- These are users who previously had short-term dating intentions that were removed.
-- "Meaningful Connection" is the most flexible option that works for people
-- who were open to short-term relationships.

UPDATE user_profiles
SET dating_intention_id = (
    SELECT id FROM dating_intentions WHERE label = 'Meaningful Connection' LIMIT 1
)
WHERE dating_intention_id IS NULL;

-- +goose Down
-- Revert: Set mapped users back to NULL
-- Note: We can't perfectly restore which users had which short-term option,
-- so we'll just set them back to NULL
UPDATE user_profiles
SET dating_intention_id = NULL
WHERE dating_intention_id = (
    SELECT id FROM dating_intentions WHERE label = 'Meaningful Connection' LIMIT 1
);

