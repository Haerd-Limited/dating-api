-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_moderation_warnings (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    report_id       UUID REFERENCES user_reports (id) ON DELETE SET NULL,
    message         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    acknowledged_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS ix_user_mod_warnings_user_unack
    ON user_moderation_warnings (user_id)
    WHERE acknowledged_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_moderation_warnings;

-- +goose StatementEnd
