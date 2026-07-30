-- +goose Up
CREATE TABLE IF NOT EXISTS daily_pick_batches (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    batch_date  DATE NOT NULL,
    state       TEXT NOT NULL DEFAULT 'pending'
                CHECK (state IN ('pending','revealed','completed')),
    revealed_at TIMESTAMPTZ,
    notified_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT daily_pick_batches_user_date_uniq UNIQUE (user_id, batch_date)
);

CREATE UNIQUE INDEX IF NOT EXISTS daily_pick_batches_one_pending
    ON daily_pick_batches (user_id) WHERE state = 'pending';
CREATE UNIQUE INDEX IF NOT EXISTS daily_pick_batches_one_revealed
    ON daily_pick_batches (user_id) WHERE state = 'revealed';

CREATE TABLE IF NOT EXISTS daily_picks (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id              UUID NOT NULL REFERENCES daily_pick_batches(id) ON DELETE CASCADE,
    user_id               UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position              SMALLINT NOT NULL,
    compatibility_percent SMALLINT NOT NULL,
    overlap_count         SMALLINT NOT NULL,
    prefs_satisfied       SMALLINT NOT NULL,
    status                TEXT NOT NULL DEFAULT 'pending'
                          CHECK (status IN ('pending','liked','passed','expired')),
    decided_at            TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT daily_picks_batch_target_uniq UNIQUE (batch_id, target_user_id),
    CONSTRAINT daily_picks_no_self CHECK (user_id <> target_user_id)
);

CREATE INDEX IF NOT EXISTS idx_daily_picks_user_status
    ON daily_picks (user_id, status);

-- +goose Down
DROP INDEX IF EXISTS idx_daily_picks_user_status;
DROP TABLE IF EXISTS daily_picks;
DROP INDEX IF EXISTS daily_pick_batches_one_revealed;
DROP INDEX IF EXISTS daily_pick_batches_one_pending;
DROP TABLE IF EXISTS daily_pick_batches;
