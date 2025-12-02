-- +goose Up
-- Add sort_order column to questions table for sequential question ordering within categories
ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS sort_order INT;

-- Set sort_order to id for existing questions (maintains current order)
UPDATE questions
SET sort_order = id
WHERE sort_order IS NULL;

-- Make sort_order NOT NULL after setting values
ALTER TABLE questions
    ALTER COLUMN sort_order SET NOT NULL;

-- Add index for efficient sequential queries
CREATE INDEX IF NOT EXISTS idx_questions_category_sort ON questions(category_id, sort_order);

-- +goose Down
DROP INDEX IF EXISTS idx_questions_category_sort;
ALTER TABLE questions
    DROP COLUMN IF EXISTS sort_order;

