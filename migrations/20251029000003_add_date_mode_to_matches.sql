-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Add date_mode field to matches table
 * ----------------------------------------------------------*/
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS date_mode BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS ix_matches_date_mode ON matches (date_mode);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE matches DROP COLUMN IF EXISTS date_mode;

-- +goose StatementEnd