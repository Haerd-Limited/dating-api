-- +goose Up
-- +goose StatementBegin
ALTER TABLE prompt_types
    ADD COLUMN IF NOT EXISTS is_core BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS core_position SMALLINT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_prompt_types_core_position
    ON prompt_types(core_position) WHERE is_core = TRUE;

ALTER TABLE prompt_types
    ADD CONSTRAINT prompt_types_core_position_chk
    CHECK ((is_core = FALSE AND core_position IS NULL)
        OR (is_core = TRUE  AND core_position IS NOT NULL));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE prompt_types DROP CONSTRAINT IF EXISTS prompt_types_core_position_chk;
DROP INDEX IF EXISTS uq_prompt_types_core_position;
ALTER TABLE prompt_types
    DROP COLUMN IF EXISTS core_position,
    DROP COLUMN IF EXISTS is_core;
-- +goose StatementEnd
