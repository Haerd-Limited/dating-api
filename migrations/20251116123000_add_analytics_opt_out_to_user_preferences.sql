-- +goose Up
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_preferences
    ADD COLUMN IF NOT EXISTS analytics_opt_out BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE IF EXISTS user_preferences
    DROP COLUMN IF EXISTS analytics_opt_out;
-- +goose StatementEnd


