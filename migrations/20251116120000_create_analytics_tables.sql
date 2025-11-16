-- +goose Up
-- +goose StatementBegin
-- Core events table (append-only)
CREATE TABLE IF NOT EXISTS events (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    occurred_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    user_id          UUID,
    session_id       TEXT,
    name             TEXT NOT NULL,
    props            JSONB NOT NULL DEFAULT '{}'::jsonb,
    version          INT NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_events_occurred_at ON events (occurred_at);
CREATE INDEX IF NOT EXISTS idx_events_user_id ON events (user_id);
CREATE INDEX IF NOT EXISTS idx_events_name ON events (name);
CREATE INDEX IF NOT EXISTS idx_events_gin_props ON events USING GIN (props);

-- Explicit feedback (NPS / free-text)
CREATE TABLE IF NOT EXISTS feedback (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID,
    channel     TEXT NOT NULL, -- in_app, email, social, support
    text        TEXT,
    rating      INT,           -- 1..10 or 1..5 depending on channel
    tags        TEXT[] DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_feedback_created_at ON feedback (created_at);
CREATE INDEX IF NOT EXISTS idx_feedback_user_id ON feedback (user_id);

-- Snapshot storage for computed insights (weekly, monthly, etc.)
CREATE TYPE insight_scope AS ENUM ('global', 'user');
CREATE TABLE IF NOT EXISTS insight_snapshots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key           TEXT NOT NULL, -- e.g. 'weekly_highlights', 'user_weekly'
    period_start  DATE NOT NULL,
    period_end    DATE NOT NULL,
    scope         insight_scope NOT NULL,
    user_id       UUID,
    payload       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT insight_snapshots_user_chk CHECK (
        (scope = 'global' AND user_id IS NULL) OR (scope = 'user' AND user_id IS NOT NULL)
    )
);
CREATE INDEX IF NOT EXISTS idx_insight_snapshots_key_period ON insight_snapshots (key, period_start, period_end);
CREATE INDEX IF NOT EXISTS idx_insight_snapshots_user ON insight_snapshots (user_id);

-- Yearly wrapped artifact per user
CREATE TABLE IF NOT EXISTS wrapped_annual (
    user_id    UUID NOT NULL,
    year       INT  NOT NULL,
    payload    JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, year)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wrapped_annual;
DROP TABLE IF EXISTS insight_snapshots;
DROP TYPE IF EXISTS insight_scope;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd


