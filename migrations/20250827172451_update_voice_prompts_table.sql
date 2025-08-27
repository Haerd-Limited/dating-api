-- +goose Up
-- +goose StatementBegin
ALTER TABLE voice_prompts
    ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN position SMALLINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE voice_prompts
DROP COLUMN IF EXISTS is_primary,
    DROP COLUMN IF EXISTS position;
-- +goose StatementEnd
