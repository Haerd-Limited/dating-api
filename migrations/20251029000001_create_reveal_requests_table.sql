-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Create reveal_requests table
 * ----------------------------------------------------------*/
CREATE TABLE IF NOT EXISTS reveal_requests (
    conversation_id UUID NOT NULL,
    initiator_id UUID NOT NULL,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    PRIMARY KEY (conversation_id),
    CONSTRAINT fk_reveal_requests_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_reveal_requests_user
        FOREIGN KEY (initiator_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT reveal_requests_status_chk
        CHECK (status IN ('pending', 'expired', 'confirmed', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS ix_reveal_requests_status ON reveal_requests (status);
CREATE INDEX IF NOT EXISTS ix_reveal_requests_expires_at ON reveal_requests (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS reveal_requests;

-- +goose StatementEnd