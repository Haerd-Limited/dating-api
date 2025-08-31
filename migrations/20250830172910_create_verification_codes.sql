-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verification_codes (
    id               BIGSERIAL PRIMARY KEY,
    channel          TEXT NOT NULL CHECK (channel IN ('email','sms')),
    identifier       TEXT NOT NULL,                        -- email or E.164 phone
    purpose          TEXT NOT NULL,                        -- e.g., 'login' | 'register' | 'reset'
    code_hash        TEXT NOT NULL,                        -- store hash, never plaintext
    expires_at       TIMESTAMPTZ NOT NULL,
    consumed_at      TIMESTAMPTZ,
    attempts         SMALLINT NOT NULL DEFAULT 0,
    max_attempts     SMALLINT NOT NULL DEFAULT 5,
    request_ip       INET,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (channel, identifier, purpose, code_hash)
    );

-- Identifier lookup
CREATE INDEX IF NOT EXISTS idx_vc_identifier ON verification_codes (identifier);

-- Active-ish lookup: only filter by immutable predicate; time check happens in WHERE
CREATE INDEX IF NOT EXISTS idx_vc_unconsumed
    ON verification_codes (identifier, channel, purpose, expires_at)
    WHERE consumed_at IS NULL;

-- Optional: speed up expiry sweeps / cron cleanup
CREATE INDEX IF NOT EXISTS idx_vc_expires_at ON verification_codes (expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_vc_expires_at;
DROP INDEX IF EXISTS idx_vc_unconsumed;
DROP INDEX IF EXISTS idx_vc_identifier;
DROP TABLE IF EXISTS verification_codes;
-- +goose StatementEnd