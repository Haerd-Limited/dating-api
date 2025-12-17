-- +goose Up
-- PostgreSQL

-- Add 'video' to verification_type enum
-- Note: ALTER TYPE ... ADD VALUE cannot be executed inside a transaction block,
-- but Goose wraps migrations in transactions, so we use a DO block workaround
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_enum WHERE enumlabel = 'video' AND enumtypid = (SELECT oid FROM pg_type WHERE typname = 'verification_type')) THEN
    ALTER TYPE verification_type ADD VALUE 'video';
  END IF;
END
$$;
-- +goose StatementEnd

-- Add new columns to verification_attempts
ALTER TABLE verification_attempts 
  ADD COLUMN IF NOT EXISTS verification_code TEXT,
  ADD COLUMN IF NOT EXISTS video_s3_key TEXT;

-- +goose Down

ALTER TABLE verification_attempts 
  DROP COLUMN IF EXISTS verification_code,
  DROP COLUMN IF EXISTS video_s3_key;

-- Note: We cannot remove the enum value 'video' from verification_type enum in PostgreSQL
-- without dropping and recreating the enum type, which would require dropping dependent columns.
-- This is typically not done in production migrations for safety.
