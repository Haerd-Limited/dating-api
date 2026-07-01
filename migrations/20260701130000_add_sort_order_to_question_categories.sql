-- +goose Up
-- Add sort_order column to question_categories so packs can be reordered from the admin UI.
ALTER TABLE question_categories
    ADD COLUMN IF NOT EXISTS sort_order INT;

-- Backfill by current id order to preserve the existing (Phase 2) predictive-strength ordering.
UPDATE question_categories qc
SET sort_order = sub.rn
FROM (SELECT id, ROW_NUMBER() OVER (ORDER BY id) AS rn FROM question_categories) sub
WHERE qc.id = sub.id
  AND qc.sort_order IS NULL;

-- Make sort_order NOT NULL after backfilling.
ALTER TABLE question_categories
    ALTER COLUMN sort_order SET NOT NULL;

-- Add index for ordered reads.
CREATE INDEX IF NOT EXISTS idx_question_categories_sort ON question_categories(sort_order);

-- +goose Down
DROP INDEX IF EXISTS idx_question_categories_sort;
ALTER TABLE question_categories
    DROP COLUMN IF EXISTS sort_order;
