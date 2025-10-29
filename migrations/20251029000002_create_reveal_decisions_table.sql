-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Create reveal_decisions table
 * ----------------------------------------------------------*/
CREATE TABLE IF NOT EXISTS reveal_decisions (
    conversation_id UUID NOT NULL,
    user_id UUID NOT NULL,
    decision TEXT NOT NULL,
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, user_id),
    CONSTRAINT fk_reveal_decisions_conversation
        FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_reveal_decisions_user
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT reveal_decisions_decision_chk
        CHECK (decision IN ('continue', 'date', 'unmatch'))
);

CREATE INDEX IF NOT EXISTS ix_reveal_decisions_conversation ON reveal_decisions (conversation_id);
CREATE INDEX IF NOT EXISTS ix_reveal_decisions_user ON reveal_decisions (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS reveal_decisions;

-- +goose StatementEnd