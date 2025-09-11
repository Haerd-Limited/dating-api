-- +goose Up
-- +goose StatementBegin
ALTER TABLE voice_prompts ADD COLUMN cover_photo_url TEXT;
ALTER TABLE user_profiles ADD COLUMN cover_photo_url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE voice_prompts
DROP COLUMN IF EXISTS cover_photo_url;

ALTER TABLE user_profiles
DROP COLUMN IF EXISTS cover_photo_url;
-- +goose StatementEnd
