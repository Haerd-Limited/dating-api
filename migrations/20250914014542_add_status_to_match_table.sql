-- +goose Up
-- +goose StatementBegin

-- 1) Create enum if it doesn't already exist
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'match_status') THEN
CREATE TYPE match_status AS ENUM ('active','unmatched','blocked');
END IF;
END$$;

-- 2) Add column to matches table
ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS status match_status NOT NULL DEFAULT 'active';

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin

-- 1) Drop the column (data loss warning)
ALTER TABLE matches DROP COLUMN IF EXISTS status;

-- 2) Drop enum if no other columns depend on it
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'match_status') THEN
DROP TYPE match_status;
END IF;
END$$;

-- +goose StatementEnd