-- +goose Up
-- +goose StatementBegin

ALTER TABLE user_profiles ADD COLUMN emoji TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE user_profiles
DROP COLUMN IF EXISTS emoji;
-- +goose StatementEnd
