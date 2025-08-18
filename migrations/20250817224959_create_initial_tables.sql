-- +goose Up
-- Core extensions
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================
-- Lookup tables
-- =========================
CREATE TABLE genders (
     id SMALLSERIAL PRIMARY KEY ,
     label TEXT UNIQUE NOT NULL
);

CREATE TABLE sexualities (
         id SMALLSERIAL PRIMARY KEY,
         label TEXT UNIQUE NOT NULL
);

CREATE TABLE relationship_types (
                id SMALLSERIAL PRIMARY KEY,
                label TEXT UNIQUE NOT NULL
);

CREATE TABLE dating_intentions (
               id SMALLSERIAL PRIMARY KEY,
               label TEXT UNIQUE NOT NULL
);

CREATE TABLE religions (
       id SMALLSERIAL PRIMARY KEY,
       label TEXT UNIQUE NOT NULL
);

CREATE TABLE education_levels (
              id SMALLSERIAL PRIMARY KEY,
              label TEXT UNIQUE NOT NULL
);

-- generic "habits" (drinking/smoking/marijuana/drugs)
CREATE TABLE habits (
    id SMALLSERIAL PRIMARY KEY,
    label TEXT UNIQUE NOT NULL
);

CREATE TABLE family_statuses (           -- e.g., "Don't have children", "Have children"
             id SMALLSERIAL PRIMARY KEY,
             label TEXT UNIQUE NOT NULL
);

CREATE TABLE family_plans (              -- e.g., "Want children", "Don't want children"
          id SMALLSERIAL PRIMARY KEY,
          label TEXT UNIQUE NOT NULL
);

CREATE TABLE ethnicities (
         id SMALLSERIAL PRIMARY KEY,
         label TEXT UNIQUE NOT NULL
);

CREATE TABLE languages (
       id SMALLSERIAL PRIMARY KEY,
       code TEXT UNIQUE,                      -- optional ISO code
       label TEXT UNIQUE NOT NULL
);

CREATE TABLE interests (
       id SMALLSERIAL PRIMARY KEY,
       label TEXT UNIQUE NOT NULL
);

CREATE TABLE prompt_types (
          id SMALLSERIAL PRIMARY KEY,
          key TEXT UNIQUE NOT NULL,              -- e.g., "best_advice", "two_truths"
          label TEXT NOT NULL
);

-- =========================
-- Users & Profiles
-- =========================
CREATE TABLE users (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   email TEXT UNIQUE,
   phone TEXT UNIQUE,
   password_hash TEXT,                    -- if applicable
   created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
   updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Visibility enum for per-field privacy settings
-- +goose StatementBegin
DO $VL$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'visibility_level' AND n.nspname = 'public'
  ) THEN
CREATE TYPE public.visibility_level AS ENUM
    ('hidden','visible','always_hidden','always_visible');
END IF;
END;
$VL$;
-- +goose StatementEnd

CREATE TABLE user_profiles (
           user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

           display_name TEXT,
           birthdate DATE,
           height_cm SMALLINT,
           bio TEXT,

-- location
           geo GEOGRAPHY(Point, 4326),
           city TEXT,
           country TEXT,

-- single-selects (FKs)
           gender_id SMALLINT REFERENCES genders(id),
           sexuality_id SMALLINT REFERENCES sexualities(id),
           relationship_type_id SMALLINT REFERENCES relationship_types(id),
           dating_intention_id SMALLINT REFERENCES dating_intentions(id),
           religion_id SMALLINT REFERENCES religions(id),
           education_level_id SMALLINT REFERENCES education_levels(id),

           drinking_id SMALLINT REFERENCES habits(id),
           smoking_id SMALLINT REFERENCES habits(id),
           marijuana_id SMALLINT REFERENCES habits(id),
           drugs_id SMALLINT REFERENCES habits(id),

           children_status_id SMALLINT REFERENCES family_statuses(id),
           family_plan_id SMALLINT REFERENCES family_plans(id),

           ethnicity_id SMALLINT REFERENCES ethnicities(id),

-- simple text fields
           work TEXT,
           job_title TEXT,
           university TEXT,

           profile_meta JSONB,                    -- long-tail extras

           created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
           updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_profiles_geo        ON user_profiles USING GIST (geo);
CREATE INDEX idx_user_profiles_birthdate  ON user_profiles (birthdate);
CREATE INDEX idx_user_profiles_gender     ON user_profiles (gender_id);
CREATE INDEX idx_user_profiles_sexuality  ON user_profiles (sexuality_id);
CREATE INDEX idx_user_profiles_reltype    ON user_profiles (relationship_type_id);
CREATE INDEX idx_user_profiles_intention  ON user_profiles (dating_intention_id);
CREATE INDEX idx_user_profiles_religion   ON user_profiles (religion_id);
CREATE INDEX idx_user_profiles_edu        ON user_profiles (education_level_id);

-- Per-field visibility map
CREATE TABLE user_profile_visibility (
                     user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                     field_name TEXT NOT NULL,                          -- e.g., 'religion_id', 'height_cm'
                     visibility visibility_level NOT NULL DEFAULT 'hidden',
                     PRIMARY KEY (user_id, field_name)
);

-- =========================
-- Preferences (what a user wants to see)
-- =========================
CREATE TABLE user_preferences (
              user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

              distance_km SMALLINT,                  -- max search radius

              age_min SMALLINT,
              age_max SMALLINT,

              seek_gender_ids INT[],                 -- arrays allow multi-acceptance
              seek_intention_ids INT[],
              seek_religion_ids INT[],
              seek_relationship_type_ids INT[],

              created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
              updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_prefs_seek_gender     ON user_preferences USING GIN (seek_gender_ids);
CREATE INDEX idx_prefs_seek_intention  ON user_preferences USING GIN (seek_intention_ids);
CREATE INDEX idx_prefs_seek_religion   ON user_preferences USING GIN (seek_religion_ids);
CREATE INDEX idx_prefs_seek_reltype    ON user_preferences USING GIN (seek_relationship_type_ids);

-- =========================
-- Multi-select joins
-- =========================
CREATE TABLE user_languages (
            user_id UUID REFERENCES users(id) ON DELETE CASCADE,
            language_id SMALLINT REFERENCES languages(id),
            PRIMARY KEY (user_id, language_id)
);
CREATE INDEX idx_user_languages_lang ON user_languages(language_id);

CREATE TABLE user_interests (
            user_id UUID REFERENCES users(id) ON DELETE CASCADE,
            interest_id SMALLINT REFERENCES interests(id),
            PRIMARY KEY (user_id, interest_id)
);
CREATE INDEX idx_user_interests_interest ON user_interests(interest_id);

-- =========================
-- Media (photos & voice prompts)
-- =========================
CREATE TABLE photos (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    position SMALLINT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    is_hidden_until_reveal BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_photos_user ON photos(user_id);

CREATE TABLE voice_prompts (
           id BIGSERIAL PRIMARY KEY,
           user_id UUID REFERENCES users(id) ON DELETE CASCADE,
           prompt_type SMALLINT REFERENCES prompt_types(id),
           audio_url TEXT NOT NULL,
           duration_ms INT NOT NULL,
           transcript TEXT,
           created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_voice_prompts_user ON voice_prompts(user_id);

-- =========================
-- Matches & messaging (minimal for v1)
-- =========================
CREATE TABLE matches (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_a UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     user_b UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
     revealed_at TIMESTAMPTZ,
     CONSTRAINT matches_distinct CHECK (user_a <> user_b)
);

-- Prevent duplicate pairs regardless of order (A,B) vs (B,A)
CREATE UNIQUE INDEX idx_matches_pair_unique
    ON matches (LEAST(user_a, user_b), GREATEST(user_a, user_b));

CREATE TABLE messages_voice (
            id BIGSERIAL PRIMARY KEY,
            match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
            sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
            audio_url TEXT NOT NULL,
            duration_ms INT NOT NULL,
            created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_messages_voice_match ON messages_voice(match_id);
CREATE INDEX idx_messages_voice_sender ON messages_voice(sender_id);

-- +goose Down
DROP TABLE IF EXISTS messages_voice;
DROP TABLE IF EXISTS matches;
DROP INDEX IF EXISTS idx_voice_prompts_user;
DROP TABLE IF EXISTS voice_prompts;
DROP INDEX IF EXISTS idx_photos_user;
DROP TABLE IF EXISTS photos;

DROP INDEX IF EXISTS idx_user_interests_interest;
DROP TABLE IF EXISTS user_interests;
DROP INDEX IF EXISTS idx_user_languages_lang;
DROP TABLE IF EXISTS user_languages;

DROP INDEX IF EXISTS idx_prefs_seek_reltype;
DROP INDEX IF EXISTS idx_prefs_seek_religion;
DROP INDEX IF EXISTS idx_prefs_seek_intention;
DROP INDEX IF EXISTS idx_prefs_seek_gender;
DROP TABLE IF EXISTS user_preferences;

DROP TABLE IF EXISTS user_profile_visibility;
DROP INDEX IF EXISTS idx_user_profiles_edu;
DROP INDEX IF EXISTS idx_user_profiles_religion;
DROP INDEX IF EXISTS idx_user_profiles_intention;
DROP INDEX IF EXISTS idx_user_profiles_reltype;
DROP INDEX IF EXISTS idx_user_profiles_sexuality;
DROP INDEX IF EXISTS idx_user_profiles_gender;
DROP INDEX IF EXISTS idx_user_profiles_birthdate;
DROP INDEX IF EXISTS idx_user_profiles_geo;
DROP TABLE IF EXISTS user_profiles;

DROP TABLE IF EXISTS prompt_types;
DROP TABLE IF EXISTS interests;
DROP TABLE IF EXISTS languages;
DROP TABLE IF EXISTS ethnicities;
DROP TABLE IF EXISTS family_plans;
DROP TABLE IF EXISTS family_statuses;
DROP TABLE IF EXISTS habits;
DROP TABLE IF EXISTS education_levels;
DROP TABLE IF EXISTS religions;
DROP TABLE IF EXISTS dating_intentions;
DROP TABLE IF EXISTS relationship_types;
DROP TABLE IF EXISTS sexualities;
DROP TABLE IF EXISTS genders;

DROP TABLE IF EXISTS users;

-- +goose StatementBegin
DO $VL$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM pg_type t
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE t.typname = 'visibility_level' AND n.nspname = 'public'
  ) THEN
DROP TYPE public.visibility_level;
END IF;
END;
$VL$;
-- +goose StatementEnd

-- Leave PostGIS extension installed; removing extensions in Down can break shared deps.