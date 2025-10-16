-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS ux_verif_attempt_user_session
    ON verification_attempts (user_id, session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS ux_verif_attempt_user_session;
-- +goose StatementEnd
