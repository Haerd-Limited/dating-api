-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_feedback_user_id;
DROP INDEX IF EXISTS idx_feedback_created_at;
DROP TABLE IF EXISTS feedback;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Recreate old feedback table for rollback
CREATE TABLE IF NOT EXISTS feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID,
    channel     TEXT NOT NULL,
    text        TEXT,
    rating      INT,
    tags        TEXT[] DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback (created_at);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON feedback (user_id);
-- +goose StatementEnd

