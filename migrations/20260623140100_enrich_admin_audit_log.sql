-- +goose Up
-- +goose StatementBegin
ALTER TABLE admin_audit_log
    ADD COLUMN IF NOT EXISTS actor_session_id UUID,
    ADD COLUMN IF NOT EXISTS actor_name TEXT;

CREATE INDEX IF NOT EXISTS idx_admin_audit_actor_name ON admin_audit_log (actor_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE admin_audit_log
    DROP COLUMN IF EXISTS actor_name,
    DROP COLUMN IF EXISTS actor_session_id;
-- +goose StatementEnd
