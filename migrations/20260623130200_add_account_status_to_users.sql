-- +goose Up
-- +goose StatementBegin

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS account_status TEXT NOT NULL DEFAULT 'active',
    ADD COLUMN IF NOT EXISTS suspended_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS moderation_reason TEXT,
    ADD COLUMN IF NOT EXISTS status_updated_at TIMESTAMPTZ;

ALTER TABLE users
    ADD CONSTRAINT users_account_status_chk
    CHECK (account_status IN ('active', 'suspended', 'banned'));

CREATE INDEX IF NOT EXISTS ix_users_account_status
    ON users (account_status)
    WHERE account_status <> 'active';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS ix_users_account_status;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_account_status_chk;

ALTER TABLE users
    DROP COLUMN IF EXISTS status_updated_at,
    DROP COLUMN IF EXISTS moderation_reason,
    DROP COLUMN IF EXISTS suspended_until,
    DROP COLUMN IF EXISTS account_status;

-- +goose StatementEnd
