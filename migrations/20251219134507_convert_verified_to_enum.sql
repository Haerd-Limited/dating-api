-- +goose Up
-- +goose StatementBegin
-- Step 1: Create enum type (drop and recreate if exists to ensure correct values)
DO $$
BEGIN
    -- Only drop if not in use
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE udt_name = 'user_video_verification_status'
    ) THEN
        DROP TYPE IF EXISTS user_video_verification_status;
    END IF;
    
    -- Create the enum type
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_video_verification_status') THEN
        CREATE TYPE user_video_verification_status AS ENUM ('VERIFIED', 'UNVERIFIED', 'UNDER_REVIEW');
    END IF;
END$$;

-- Step 2: Check if migration already completed
DO $$
DECLARE
    col_exists boolean;
    is_enum_type boolean;
BEGIN
    -- Check if verified column exists and is already enum type
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'verified'
    ) INTO col_exists;
    
    IF col_exists THEN
        SELECT EXISTS (
            SELECT 1 FROM information_schema.columns 
            WHERE table_schema = 'public'
            AND table_name = 'user_profiles' 
            AND column_name = 'verified'
            AND udt_name = 'user_video_verification_status'
        ) INTO is_enum_type;
        
        -- If already enum type, just set constraints and exit
        IF is_enum_type THEN
            BEGIN
                ALTER TABLE user_profiles
                    ALTER COLUMN verified SET NOT NULL,
                    ALTER COLUMN verified SET DEFAULT 'UNVERIFIED'::user_video_verification_status;
            EXCEPTION WHEN OTHERS THEN
                NULL; -- Constraints might already be set
            END;
            RETURN;
        END IF;
    END IF;
END$$;

-- Step 3: Add temporary text column for migration
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'verified_new'
    ) THEN
        ALTER TABLE user_profiles ADD COLUMN verified_new text;
    END IF;
END$$;

-- Step 4: Migrate data to text column
DO $$
BEGIN
    -- Check if old boolean column exists
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'verified'
        AND data_type = 'boolean'
    ) THEN
        -- Migrate from boolean
        UPDATE user_profiles
        SET verified_new = CASE
            WHEN verified = true THEN 'VERIFIED'
            ELSE 'UNVERIFIED'
        END;
    ELSE
        -- No boolean column, set default
        UPDATE user_profiles
        SET verified_new = 'UNVERIFIED'
        WHERE verified_new IS NULL;
    END IF;
END$$;

-- Step 5: Convert text column to enum
ALTER TABLE user_profiles
    ALTER COLUMN verified_new TYPE user_video_verification_status 
    USING verified_new::user_video_verification_status;

-- Step 6: Set constraints
ALTER TABLE user_profiles
    ALTER COLUMN verified_new SET NOT NULL,
    ALTER COLUMN verified_new SET DEFAULT 'UNVERIFIED'::user_video_verification_status;

-- Step 7: Drop old column if it exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'verified'
        AND data_type = 'boolean'
    ) THEN
        ALTER TABLE user_profiles DROP COLUMN verified;
    END IF;
END$$;

-- Step 8: Rename new column to verified
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_schema = 'public'
        AND table_name = 'user_profiles' 
        AND column_name = 'verified_new'
    ) THEN
        ALTER TABLE user_profiles RENAME COLUMN verified_new TO verified;
    END IF;
END$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Rename verified back to verified_new
ALTER TABLE user_profiles
    RENAME COLUMN verified TO verified_new;

-- Add back the boolean column
ALTER TABLE user_profiles
    ADD COLUMN verified bool not null default false;

-- Migrate data back: VERIFIED -> true, everything else -> false
UPDATE user_profiles
SET verified = CASE
    WHEN verified_new::text = 'VERIFIED' THEN true
    ELSE false
END;

-- Drop the enum column
ALTER TABLE user_profiles
    DROP COLUMN verified_new;

-- Drop the enum type
DROP TYPE IF EXISTS user_video_verification_status;
-- +goose StatementEnd
