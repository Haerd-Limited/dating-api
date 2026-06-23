-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_attempts
    ADD COLUMN IF NOT EXISTS reviewed_by_name TEXT,
    ADD COLUMN IF NOT EXISTS reviewed_by_session_id UUID,
    ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS review_notes TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE verification_attempts
    DROP COLUMN IF EXISTS review_notes,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by_session_id,
    DROP COLUMN IF EXISTS reviewed_by_name;
-- +goose StatementEnd
