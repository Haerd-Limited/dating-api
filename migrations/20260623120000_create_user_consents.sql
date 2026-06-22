-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_consents (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type TEXT NOT NULL CHECK (consent_type IN ('privacy_policy','terms_of_service')),
    version      TEXT NOT NULL,
    accepted     BOOLEAN NOT NULL,
    accepted_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at   TIMESTAMPTZ,
    ip           INET,
    user_agent   TEXT,
    UNIQUE (user_id, consent_type, version)
);

CREATE INDEX IF NOT EXISTS idx_user_consents_user_type ON user_consents (user_id, consent_type) WHERE revoked_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_consents;
-- +goose StatementEnd
