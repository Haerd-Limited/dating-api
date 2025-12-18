-- +goose Up
-- PostgreSQL

-- Remove columns that were added via migrations that were later deleted
-- These columns exist in the database but have no corresponding migration files
ALTER TABLE verification_attempts 
  DROP COLUMN IF EXISTS challenge_code,
  DROP COLUMN IF EXISTS submitted_at,
  DROP COLUMN IF EXISTS reviewed_at,
  DROP COLUMN IF EXISTS review_notes;

-- +goose Down

-- Restore the columns if rollback is needed
ALTER TABLE verification_attempts 
  ADD COLUMN IF NOT EXISTS challenge_code TEXT,
  ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS review_notes TEXT;
