-- +goose Up
-- +goose StatementBegin
-- Step 1: Create enum type for cover_media_type
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'cover_media_type') THEN
        CREATE TYPE cover_media_type AS ENUM ('image', 'gif');
    END IF;
END$$;

-- Step 2: Rename cover_photo_url to cover_media_url in voice_prompts
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'voice_prompts' 
        AND column_name = 'cover_photo_url'
    ) THEN
        ALTER TABLE voice_prompts RENAME COLUMN cover_photo_url TO cover_media_url;
    END IF;
END$$;

-- Step 3: Rename cover_photo_url to cover_media_url in user_profiles
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'cover_photo_url'
    ) THEN
        ALTER TABLE user_profiles RENAME COLUMN cover_photo_url TO cover_media_url;
    END IF;
END$$;

-- Step 4: Add cover_media_type column to voice_prompts
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'voice_prompts' 
        AND column_name = 'cover_media_type'
    ) THEN
        ALTER TABLE voice_prompts ADD COLUMN cover_media_type cover_media_type;
    END IF;
END$$;

-- Step 5: Add cover_media_aspect_ratio column to voice_prompts
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'voice_prompts' 
        AND column_name = 'cover_media_aspect_ratio'
    ) THEN
        ALTER TABLE voice_prompts ADD COLUMN cover_media_aspect_ratio REAL;
    END IF;
END$$;

-- Step 6: Add cover_media_type column to user_profiles
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'cover_media_type'
    ) THEN
        ALTER TABLE user_profiles ADD COLUMN cover_media_type cover_media_type;
    END IF;
END$$;

-- Step 7: Add cover_media_aspect_ratio column to user_profiles
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'cover_media_aspect_ratio'
    ) THEN
        ALTER TABLE user_profiles ADD COLUMN cover_media_aspect_ratio REAL;
    END IF;
END$$;

-- Step 8: Migrate existing data - set cover_media_type to 'image' for existing URLs
UPDATE voice_prompts
SET cover_media_type = 'image'::cover_media_type
WHERE cover_media_url IS NOT NULL 
  AND cover_media_type IS NULL;

UPDATE user_profiles
SET cover_media_type = 'image'::cover_media_type
WHERE cover_media_url IS NOT NULL 
  AND cover_media_type IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Step 1: Remove data migration (no need to revert data)
-- Step 2: Drop new columns from user_profiles
ALTER TABLE user_profiles
    DROP COLUMN IF EXISTS cover_media_aspect_ratio;
ALTER TABLE user_profiles
    DROP COLUMN IF EXISTS cover_media_type;

-- Step 3: Drop new columns from voice_prompts
ALTER TABLE voice_prompts
    DROP COLUMN IF EXISTS cover_media_aspect_ratio;
ALTER TABLE voice_prompts
    DROP COLUMN IF EXISTS cover_media_type;

-- Step 4: Rename cover_media_url back to cover_photo_url in user_profiles
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'cover_media_url'
    ) THEN
        ALTER TABLE user_profiles RENAME COLUMN cover_media_url TO cover_photo_url;
    END IF;
END$$;

-- Step 5: Rename cover_media_url back to cover_photo_url in voice_prompts
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'voice_prompts' 
        AND column_name = 'cover_media_url'
    ) THEN
        ALTER TABLE voice_prompts RENAME COLUMN cover_media_url TO cover_photo_url;
    END IF;
END$$;

-- Step 6: Drop enum type (only if not in use)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE udt_name = 'cover_media_type'
    ) THEN
        DROP TYPE IF EXISTS cover_media_type;
    END IF;
END$$;
-- +goose StatementEnd

