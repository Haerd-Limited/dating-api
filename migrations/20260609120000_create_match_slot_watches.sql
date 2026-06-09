-- +goose Up
-- +goose StatementBegin
CREATE TABLE match_slot_watches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    watcher_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    watched_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT match_slot_watches_distinct CHECK (watcher_user_id <> watched_user_id),
    UNIQUE (watcher_user_id, watched_user_id)
);

CREATE INDEX idx_match_slot_watches_watched ON match_slot_watches (watched_user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS match_slot_watches;
-- +goose StatementEnd
