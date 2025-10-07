-- +goose Up
-- +goose StatementBegin
BEGIN;

-- Create participant rows for user_a
INSERT INTO conversation_participants (conversation_id, user_id, score, score_lifetime, last_contrib_at)
SELECT c.id, c.user_a, 0, 0, NULL
FROM conversations c
    ON CONFLICT (conversation_id, user_id) DO NOTHING;

-- Create participant rows for user_b
INSERT INTO conversation_participants (conversation_id, user_id, score, score_lifetime, last_contrib_at)
SELECT c.id, c.user_b, 0, 0, NULL
FROM conversations c
    ON CONFLICT (conversation_id, user_id) DO NOTHING;

COMMIT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
BEGIN;

-- Remove only rows that correspond to (conversation.user_a or conversation.user_b)
DELETE FROM conversation_participants cp
    USING conversations c
WHERE cp.conversation_id = c.id
  AND (cp.user_id = c.user_a OR cp.user_id = c.user_b);

COMMIT;
-- +goose StatementEnd
