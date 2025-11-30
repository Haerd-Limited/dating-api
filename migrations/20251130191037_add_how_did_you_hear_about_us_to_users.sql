-- +goose Up
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS how_did_you_hear_about_us TEXT;

-- +goose Down
ALTER TABLE users
    DROP COLUMN IF EXISTS how_did_you_hear_about_us;

