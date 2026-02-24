-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS data_export_requests (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requested_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_data_export_requests_user_id ON data_export_requests (user_id);
CREATE INDEX IF NOT EXISTS idx_data_export_requests_requested_at ON data_export_requests (requested_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS data_export_requests;
-- +goose StatementEnd
