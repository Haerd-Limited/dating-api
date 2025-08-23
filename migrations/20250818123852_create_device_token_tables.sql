-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS device_tokens (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                               token TEXT NOT NULL,
                               created_at TIMESTAMP DEFAULT now(),
                               updated_at TIMESTAMP DEFAULT now(),
                               UNIQUE(user_id, token)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE device_tokens;
-- +goose StatementEnd
