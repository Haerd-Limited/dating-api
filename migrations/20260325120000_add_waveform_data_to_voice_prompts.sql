-- +goose Up
-- +goose StatementBegin
ALTER TABLE voice_prompts
    ADD COLUMN waveform_data JSONB;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE voice_prompts
    DROP COLUMN IF EXISTS waveform_data;
-- +goose StatementEnd
