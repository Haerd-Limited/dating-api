-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS voice_worth_hearing_weekly (
    user_id       UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    week_start    DATE        NOT NULL,
    candidate_ids UUID[]      NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, week_start)
);

CREATE INDEX IF NOT EXISTS idx_voice_worth_hearing_weekly_week_start
    ON voice_worth_hearing_weekly (week_start);

DROP TRIGGER IF EXISTS trg_set_updated_at_vwh_weekly ON voice_worth_hearing_weekly;
CREATE TRIGGER trg_set_updated_at_vwh_weekly
    BEFORE UPDATE ON voice_worth_hearing_weekly
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_set_updated_at_vwh_weekly ON voice_worth_hearing_weekly;
DROP INDEX IF EXISTS idx_voice_worth_hearing_weekly_week_start;
DROP TABLE IF EXISTS voice_worth_hearing_weekly;
-- +goose StatementEnd

