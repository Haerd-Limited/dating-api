-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS swipes (
    id               BIGSERIAL PRIMARY KEY,

    actor_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- what the actor did to the target
    action           TEXT NOT NULL CHECK (action IN ('like','pass','superlike')),

    -- optional idempotency key from client; unique per actor when present
    idempotency_key  TEXT,

    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- don’t allow self-swipes
    CONSTRAINT swipes_no_self CHECK (actor_id <> target_id),

    -- one row per (actor → target)
    CONSTRAINT swipes_actor_target_uniq UNIQUE (actor_id, target_id)
    );

-- unique per (actor, idempotency_key) when a key is provided
CREATE UNIQUE INDEX IF NOT EXISTS idx_swipes_actor_idem_unique
    ON swipes (actor_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

-- fast reciprocal-like lookup (did target already like actor?)
CREATE INDEX IF NOT EXISTS idx_swipes_target_actor_like
    ON swipes (target_id, actor_id)
    WHERE action IN ('like','superlike');

-- helpful filter for a user's outgoing swipes
CREATE INDEX IF NOT EXISTS idx_swipes_actor ON swipes (actor_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_swipes_actor;
DROP INDEX IF EXISTS idx_swipes_target_actor_like;
DROP INDEX IF EXISTS idx_swipes_actor_idem_unique;
DROP TABLE IF EXISTS swipes;
-- +goose StatementEnd
