-- +goose Up
-- PostgreSQL

-- If you prefer pgcrypto's gen_random_uuid(), enable this instead of uuid-ossp:
CREATE EXTENSION IF NOT EXISTS "pgcrypto";


-- 1) Enums (wrap DO block so Goose doesn't split it)
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'verification_type') THEN
CREATE TYPE verification_type AS ENUM ('photo');
END IF;

  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'verification_status') THEN
CREATE TYPE verification_status AS ENUM ('pending','passed','failed','needs_review');
END IF;
END
$$;
-- +goose StatementEnd

-- 2) Attempts table
CREATE TABLE IF NOT EXISTS verification_attempts (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL,
    type               verification_type NOT NULL DEFAULT 'photo',
    status             verification_status NOT NULL DEFAULT 'pending',
    session_id         TEXT,                    -- Rekognition Face Liveness session id
    liveness_score     DOUBLE PRECISION,        -- confidence
    match_score        DOUBLE PRECISION,        -- best Rekognition similarity (0-100)
    reason_codes       JSONB,                   -- array/obj with reasons
    best_frame_s3_key  TEXT,                    -- optional: where best frame is stored
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_verif_attempt_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS idx_verif_attempts_user_created
    ON verification_attempts (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_verif_attempts_status
    ON verification_attempts (status);

CREATE INDEX IF NOT EXISTS idx_verif_attempts_session
    ON verification_attempts (session_id);

-- 3) User verification status table
CREATE TABLE IF NOT EXISTS user_verification_status (
    user_id            UUID PRIMARY KEY,
    photo_verified     BOOLEAN NOT NULL DEFAULT FALSE,
    photo_verified_at  TIMESTAMPTZ,
    last_attempt_id    UUID,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT fk_uvs_user
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_uvs_attempt
    FOREIGN KEY (last_attempt_id) REFERENCES verification_attempts(id) ON DELETE SET NULL
    );

-- 4) updated_at helper + triggers
-- If you don't want automatic updated_at, you can delete this whole section.

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at_timestamp()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_set_updated_at_verif_attempts ON verification_attempts;
CREATE TRIGGER trg_set_updated_at_verif_attempts
    BEFORE UPDATE ON verification_attempts
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at_timestamp();

DROP TRIGGER IF EXISTS trg_set_updated_at_uvs ON user_verification_status;
CREATE TRIGGER trg_set_updated_at_uvs
    BEFORE UPDATE ON user_verification_status
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at_timestamp();

-- +goose Down

DROP TRIGGER IF EXISTS trg_set_updated_at_uvs ON user_verification_status;
DROP TRIGGER IF EXISTS trg_set_updated_at_verif_attempts ON verification_attempts;

-- Only drop the function if nothing else depends on it
DROP FUNCTION IF EXISTS set_updated_at_timestamp;

DROP TABLE IF EXISTS user_verification_status;
DROP TABLE IF EXISTS verification_attempts;

-- Drop enums inside a DO block (wrap for Goose)
-- +goose StatementBegin
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'verification_status') THEN
DROP TYPE verification_status;
END IF;

  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'verification_type') THEN
DROP TYPE verification_type;
END IF;
END
$$;
-- +goose StatementEnd

-- If you created uuid-ossp just for this project and want to remove it on down:
-- DROP EXTENSION IF EXISTS "uuid-ossp";