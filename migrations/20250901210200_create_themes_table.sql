-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_theme (
    user_id    UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    base_hex   TEXT NOT NULL,        -- "#095874"
    palette    JSONB NOT NULL,       -- ["#f3f8fa", "#d7e8f0", ..., "#02161b"]
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_theme_updated_at ON user_theme(updated_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_theme;
-- +goose StatementEnd
