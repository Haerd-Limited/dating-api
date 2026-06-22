-- +goose Up
-- +goose StatementBegin
ALTER TABLE swipes ADD COLUMN media_seconds NUMERIC(6,2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swipes DROP COLUMN IF EXISTS media_seconds;
-- +goose StatementEnd
