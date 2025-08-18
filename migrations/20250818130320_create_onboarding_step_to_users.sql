-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS onboarding_step SMALLINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE users
DROP COLUMN IF EXISTS onboarding_step;