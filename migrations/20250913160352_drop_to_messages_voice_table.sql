-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS messages_voice;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE messages_voice (
                                id BIGSERIAL PRIMARY KEY,
                                match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
                                sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                audio_url TEXT NOT NULL,
                                duration_ms INT NOT NULL,
                                created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_voice_match  ON messages_voice(match_id);
CREATE INDEX idx_messages_voice_sender ON messages_voice(sender_id);
-- +goose StatementEnd
