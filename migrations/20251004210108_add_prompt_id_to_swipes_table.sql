-- +goose Up
-- +goose StatementBegin
ALTER TABLE swipes
    ADD COLUMN prompt_id BIGINT;  -- nullable, no default
-- Add FK after column exists (NULLs are allowed under FK)
ALTER TABLE swipes
    ADD CONSTRAINT swipes_prompt_id_fkey
        FOREIGN KEY (prompt_id) REFERENCES voice_prompts(id)
            ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE swipes DROP CONSTRAINT IF EXISTS swipes_prompt_id_fkey;
ALTER TABLE swipes DROP COLUMN IF EXISTS prompt_id;
-- +goose StatementEnd
