-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_profiles ALTER COLUMN display_name SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_profiles ALTER COLUMN display_name DROP NOT NULL;
-- +goose StatementEnd
