-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
ALTER TABLE swipes
ADD COLUMN message_type TEXT,
ADD COLUMN message TEXT,
ADD COLUMN voicenote_url TEXT;

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swipes
DROP COLUMN IF EXISTS message_type,
DROP COLUMN IF EXISTS message,
DROP COLUMN IF EXISTS voicenote_url;

-- +goose StatementEnd
