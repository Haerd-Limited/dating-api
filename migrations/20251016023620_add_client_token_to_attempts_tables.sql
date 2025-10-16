-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_attempts
    ADD COLUMN client_token TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE verification_attempts
DROP COLUMN IF EXISTS client_token;
-- +goose StatementEnd
