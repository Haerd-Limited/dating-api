-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_profiles
    ADD COLUMN verified bool not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_profiles
DROP COLUMN IF EXISTS verified;
-- +goose StatementEnd
