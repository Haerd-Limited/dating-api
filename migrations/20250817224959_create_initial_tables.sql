-- +goose Up
-- Core extensions
CREATE EXTENSION IF NOT EXISTS "postgis";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================
-- Lookup tables
-- =========================
CREATE TABLE genders (
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

CREATE TABLE habits (
                        id SMALLSERIAL PRIMARY KEY,
                        label TEXT UNIQUE NOT NULL
);

CREATE TABLE family_statuses (
                                 id SMALLSERIAL PRIMARY KEY,
                                 label TEXT UNIQUE NOT NULL
);

CREATE TABLE family_plans (
                              id SMALLSERIAL PRIMARY KEY,
                              label TEXT UNIQUE NOT NULL
);

CREATE TABLE ethnicities (
                             id SMALLSERIAL PRIMARY KEY,
                             label TEXT UNIQUE NOT NULL
);

CREATE TABLE languages (
                           id SMALLSERIAL PRIMARY KEY,
                           code TEXT UNIQUE,
                           label TEXT UNIQUE NOT NULL
);

CREATE TABLE interests (
                           id SMALLSERIAL PRIMARY KEY,
                           label TEXT UNIQUE NOT NULL
);

CREATE TABLE prompt_types (
                              id SMALLSERIAL PRIMARY KEY,
                              key TEXT UNIQUE NOT NULL,
                              label TEXT NOT NULL
);

-- NEW: political beliefs
CREATE TABLE political_beliefs (
                                   id SMALLSERIAL PRIMARY KEY,
                                   label TEXT UNIQUE NOT NULL
);

-- =========================
-- Users & Profiles
-- =========================
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email TEXT UNIQUE,
                       first_name TEXT NOT NULL,
                       last_name TEXT DEFAULT '',
                       phone TEXT UNIQUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Visibility enum (wrapped so Goose doesn't split the DO block)
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
                               dating_intention_id SMALLINT REFERENCES dating_intentions(id),
                               religion_id SMALLINT REFERENCES religions(id),
                               education_level_id SMALLINT REFERENCES education_levels(id),
                               political_belief_id SMALLINT REFERENCES political_beliefs(id),

    -- habits / family
                               drinking_id SMALLINT REFERENCES habits(id),
                               smoking_id SMALLINT REFERENCES habits(id),
                               marijuana_id SMALLINT REFERENCES habits(id),
                               drugs_id SMALLINT REFERENCES habits(id),

                               children_status_id SMALLINT REFERENCES family_statuses(id),
                               family_plan_id SMALLINT REFERENCES family_plans(id),

                               ethnicity_id SMALLINT REFERENCES ethnicities(id),

    -- simple text
                               work TEXT,
                               job_title TEXT,
                               university TEXT,

                               profile_meta JSONB,

                               created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                               updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_profiles_geo        ON user_profiles USING GIST (geo);
CREATE INDEX idx_user_profiles_birthdate  ON user_profiles (birthdate);
CREATE INDEX idx_user_profiles_gender     ON user_profiles (gender_id);
CREATE INDEX idx_user_profiles_intention  ON user_profiles (dating_intention_id);
CREATE INDEX idx_user_profiles_religion   ON user_profiles (religion_id);
CREATE INDEX idx_user_profiles_edu        ON user_profiles (education_level_id);
CREATE INDEX idx_user_profiles_politics   ON user_profiles (political_belief_id);

-- Per-field visibility map
CREATE TABLE user_profile_visibility (
                                         user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                         field_name TEXT NOT NULL,
                                         visibility visibility_level NOT NULL DEFAULT 'hidden',
                                         PRIMARY KEY (user_id, field_name)
);

-- =========================
-- Preferences (what a user wants to see)
-- =========================
CREATE TABLE user_preferences (
                                  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,

                                  distance_km SMALLINT,

                                  age_min SMALLINT,
                                  age_max SMALLINT,

                                  seek_gender_ids INT[],
                                  seek_intention_ids INT[],
                                  seek_religion_ids INT[],
                                  seek_political_belief_ids INT[],   -- NEW (optional filter)

                                  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_prefs_seek_gender     ON user_preferences USING GIN (seek_gender_ids);
CREATE INDEX idx_prefs_seek_intention  ON user_preferences USING GIN (seek_intention_ids);
CREATE INDEX idx_prefs_seek_religion   ON user_preferences USING GIN (seek_religion_ids);
CREATE INDEX idx_prefs_seek_politics   ON user_preferences USING GIN (seek_political_belief_ids);

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
-- Matches & messaging
-- =========================
CREATE TABLE matches (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         user_a UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         user_b UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                         created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                         revealed_at TIMESTAMPTZ,
                         CONSTRAINT matches_distinct CHECK (user_a <> user_b)
);

-- unordered pair uniqueness (A,B) == (B,A)
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
CREATE INDEX idx_messages_voice_match  ON messages_voice(match_id);
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

DROP INDEX IF EXISTS idx_prefs_seek_politics;
DROP INDEX IF EXISTS idx_prefs_seek_religion;
DROP INDEX IF EXISTS idx_prefs_seek_intention;
DROP INDEX IF EXISTS idx_prefs_seek_gender;
DROP TABLE IF EXISTS user_preferences;

DROP TABLE IF EXISTS user_profile_visibility;

DROP INDEX IF EXISTS idx_user_profiles_politics;
DROP INDEX IF EXISTS idx_user_profiles_edu;
DROP INDEX IF EXISTS idx_user_profiles_religion;
DROP INDEX IF EXISTS idx_user_profiles_intention;
DROP INDEX IF EXISTS idx_user_profiles_gender;
DROP INDEX IF EXISTS idx_user_profiles_birthdate;
DROP INDEX IF EXISTS idx_user_profiles_geo;
DROP TABLE IF EXISTS user_profiles;

DROP TABLE IF EXISTS political_beliefs;
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
DROP TABLE IF EXISTS genders;

DROP TABLE IF EXISTS users;

-- drop the enum last (wrapped for Goose)
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