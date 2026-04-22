-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS user_theme;

ALTER TABLE user_profiles
    DROP COLUMN IF EXISTS cover_media_aspect_ratio,
    DROP COLUMN IF EXISTS cover_media_type,
    DROP COLUMN IF EXISTS cover_media_url;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cover_media_type') THEN
        CREATE TYPE cover_media_type AS ENUM ('image', 'gif');
    END IF;
END$$;

CREATE TABLE IF NOT EXISTS user_theme (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    base_hex   TEXT NOT NULL,
    palette    JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_theme_updated_at ON user_theme(updated_at);

ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS cover_media_url TEXT,
    ADD COLUMN IF NOT EXISTS cover_media_type cover_media_type,
    ADD COLUMN IF NOT EXISTS cover_media_aspect_ratio REAL;
-- +goose StatementEnd
