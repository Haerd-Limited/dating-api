-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Safety tables: user_blocks, user_reports, report_actions
 * ----------------------------------------------------------*/

CREATE TABLE IF NOT EXISTS user_blocks (
    blocker_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    blocked_user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reason          TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_user_id, blocked_user_id),
    CONSTRAINT user_blocks_self_chk CHECK (blocker_user_id <> blocked_user_id)
);

CREATE INDEX IF NOT EXISTS ix_user_blocks_blocked_user ON user_blocks (blocked_user_id);

DO $REPORT_STATUS$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'report_status') THEN
        CREATE TYPE report_status AS ENUM ('pending', 'in_review', 'resolved', 'escalated', 'dismissed');
    END IF;
END;
$REPORT_STATUS$;

CREATE TABLE IF NOT EXISTS user_reports (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_user_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    reported_user_id  UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    subject_type      TEXT NOT NULL,
    subject_id        UUID,
    category          TEXT NOT NULL,
    notes             TEXT,
    status            report_status NOT NULL DEFAULT 'pending',
    severity          TEXT,
    metadata          JSONB,
    auto_action       TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    resolved_at       TIMESTAMPTZ,
    CONSTRAINT user_reports_subject_chk CHECK (subject_type IN ('user','message','profile'))
);

CREATE INDEX IF NOT EXISTS ix_user_reports_reported_user ON user_reports (reported_user_id);
CREATE INDEX IF NOT EXISTS ix_user_reports_reporter_user ON user_reports (reporter_user_id);
CREATE INDEX IF NOT EXISTS ix_user_reports_status ON user_reports (status);
CREATE INDEX IF NOT EXISTS ix_user_reports_created_at ON user_reports (created_at DESC);

CREATE TABLE IF NOT EXISTS report_actions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id      UUID NOT NULL REFERENCES user_reports (id) ON DELETE CASCADE,
    reviewer_id    UUID REFERENCES users (id) ON DELETE SET NULL,
    action_type    TEXT NOT NULL,
    action_payload JSONB,
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ix_report_actions_report_id ON report_actions (report_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS report_actions;

DROP TABLE IF EXISTS user_reports;

DO $DROP_REPORT_STATUS$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_type
        WHERE typname = 'report_status'
    ) THEN
        DROP TYPE report_status;
    END IF;
END;
$DROP_REPORT_STATUS$;

DROP TABLE IF EXISTS user_blocks;

-- +goose StatementEnd

