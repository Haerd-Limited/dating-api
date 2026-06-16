-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sms_broadcasts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone      TEXT NOT NULL,
    message    TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('sent', 'failed')),
    error      TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sms_broadcasts_user_id ON sms_broadcasts (user_id);
CREATE INDEX IF NOT EXISTS idx_sms_broadcasts_created_at ON sms_broadcasts (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sms_broadcasts;
-- +goose StatementEnd
