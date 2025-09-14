-- +goose Up
-- +goose StatementBegin

-- 1) Enum
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'message_type') THEN
CREATE TYPE message_type AS ENUM ('text','voice','system');
END IF;
END$$;

-- 2) Threads
CREATE TABLE IF NOT EXISTS threads (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_activity_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- per-user thread state (MVP)
    a_last_read_id    BIGINT,
    b_last_read_id    BIGINT,
    a_muted           BOOLEAN NOT NULL DEFAULT FALSE,
    b_muted           BOOLEAN NOT NULL DEFAULT FALSE,

    -- for quick list views
    last_message_id   BIGINT,

    CONSTRAINT threads_distinct_users CHECK (user_a <> user_b)
    );

-- Enforce one thread per unordered pair via UNIQUE INDEX on expressions
CREATE UNIQUE INDEX IF NOT EXISTS ux_threads_pair
    ON threads (LEAST(user_a,user_b), GREATEST(user_a,user_b));

-- Optional helper (non-unique) was redundant; removed.

-- 3) Messages
CREATE TABLE IF NOT EXISTS messages (
                                        id              BIGINT PRIMARY KEY,  -- snowflake-style id generated in app
                                        thread_id       UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    sender_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type            message_type NOT NULL,

    -- text payload (for text or system msgs)
    text_body       TEXT,

    -- voice payload
    media_key       TEXT,                -- e.g. "voice_notes/threads/{thread_id}/{id}.m4a"
    media_seconds   NUMERIC(6,2),        -- duration in seconds

    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    edited_at       TIMESTAMPTZ,
    deleted_at      TIMESTAMPTZ,

    CONSTRAINT messages_payload_check CHECK (
(type = 'text'   AND text_body IS NOT NULL AND media_key IS NULL)
    OR (type = 'voice'  AND media_key IS NOT NULL)
    OR (type = 'system' AND text_body IS NOT NULL)
    )
    );

CREATE INDEX IF NOT EXISTS idx_messages_thread_order ON messages (thread_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_messages_sender_thread ON messages (sender_id, thread_id, id DESC);

-- 4) Receipts
CREATE TABLE IF NOT EXISTS message_receipts (
                                                message_id   BIGINT NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id      UUID   NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       SMALLINT NOT NULL CHECK (status IN (1,2)),  -- 1=delivered, 2=read
    at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (message_id, user_id, status)
    );

CREATE INDEX IF NOT EXISTS idx_receipts_user_status_time ON message_receipts (user_id, status, at DESC);

-- 5) FK to last message (deferrable avoids create-time cycles)
ALTER TABLE threads
    ADD CONSTRAINT threads_last_message_fk
        FOREIGN KEY (last_message_id)
            REFERENCES messages(id)
            DEFERRABLE INITIALLY DEFERRED;

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin
ALTER TABLE threads DROP CONSTRAINT IF EXISTS threads_last_message_fk;

DROP TABLE IF EXISTS message_receipts;
DROP INDEX IF EXISTS idx_messages_sender_thread;
DROP INDEX IF EXISTS idx_messages_thread_order;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS ux_threads_pair;
DROP TABLE IF EXISTS threads;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'message_type') THEN
DROP TYPE message_type;
END IF;
END$$;

-- +goose StatementEnd