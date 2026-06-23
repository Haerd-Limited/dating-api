-- +goose Up
-- +goose StatementBegin
ALTER TABLE report_actions
    ADD COLUMN IF NOT EXISTS reviewer_name TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE report_actions DROP COLUMN IF EXISTS reviewer_name;
-- +goose StatementEnd
