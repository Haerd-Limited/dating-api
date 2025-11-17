-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * Add is_active field to voice_prompts table
 * This allows us to soft-delete prompts instead of hard-deleting them,
 * preserving prompts that are referenced in conversations/swipes
 * ----------------------------------------------------------*/
ALTER TABLE voice_prompts
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;

-- Set all existing prompts to active
UPDATE voice_prompts SET is_active = TRUE WHERE is_active IS NULL;

-- Create index for better query performance when filtering by is_active
CREATE INDEX IF NOT EXISTS idx_voice_prompts_user_active ON voice_prompts(user_id, is_active) WHERE is_active = TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_voice_prompts_user_active;
ALTER TABLE voice_prompts DROP COLUMN IF EXISTS is_active;

-- +goose StatementEnd

