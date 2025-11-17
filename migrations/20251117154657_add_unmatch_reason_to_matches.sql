-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Add unmatch_reason field to matches table
 * ----------------------------------------------------------*/
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS unmatch_reason TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE matches DROP COLUMN IF EXISTS unmatch_reason;

-- +goose StatementEnd

