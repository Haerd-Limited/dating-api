-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_ip    INET,
    token_fp    TEXT,
    method      TEXT NOT NULL,
    path        TEXT NOT NULL,
    target_id   TEXT,
    status_code INT
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_occurred_at ON admin_audit_log (occurred_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS admin_audit_log;
-- +goose StatementEnd
