-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS onboarding_step TEXT NOT NULL DEFAULT 'NONE';

-- +goose Down
ALTER TABLE users
DROP COLUMN IF EXISTS onboarding_step;