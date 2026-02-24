-- +goose Up
-- +goose StatementBegin
-- Remove orphan rows so ADD CONSTRAINT does not fail (user_id not in users).
DELETE FROM events WHERE user_id IS NOT NULL AND user_id NOT IN (SELECT id FROM users);
DELETE FROM insight_snapshots WHERE user_id IS NOT NULL AND user_id NOT IN (SELECT id FROM users);
DELETE FROM wrapped_annual WHERE user_id NOT IN (SELECT id FROM users);

-- CASCADE delete analytics/insights when user is deleted (GDPR Right to Erasure).
ALTER TABLE events
    ADD CONSTRAINT fk_events_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE insight_snapshots
    ADD CONSTRAINT fk_insight_snapshots_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE wrapped_annual
    ADD CONSTRAINT fk_wrapped_annual_user_id
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events DROP CONSTRAINT IF EXISTS fk_events_user_id;
ALTER TABLE insight_snapshots DROP CONSTRAINT IF EXISTS fk_insight_snapshots_user_id;
ALTER TABLE wrapped_annual DROP CONSTRAINT IF EXISTS fk_wrapped_annual_user_id;
-- +goose StatementEnd
