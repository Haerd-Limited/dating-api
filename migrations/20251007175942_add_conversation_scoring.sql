-- +goose Up
-- +goose StatementBegin

/* ------------------------------------------------------------
 * 1) Conversations: visibility + reveal timestamp
 * ----------------------------------------------------------*/
ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS visibility_state TEXT NOT NULL DEFAULT 'HIDDEN',
    ADD COLUMN IF NOT EXISTS reveal_at TIMESTAMPTZ;

ALTER TABLE conversations
    ADD CONSTRAINT conversations_visibility_state_chk
        CHECK (visibility_state IN ('HIDDEN','REVEALED'));


/* ------------------------------------------------------------
 * 2) Per-participant progress
 * ----------------------------------------------------------*/
CREATE TABLE IF NOT EXISTS conversation_participants (
     conversation_id  UUID NOT NULL,
     user_id          UUID NOT NULL,
     score            INTEGER NOT NULL DEFAULT 0,
     score_lifetime   INTEGER NOT NULL DEFAULT 0,
     last_contrib_at  TIMESTAMPTZ,
     PRIMARY KEY (conversation_id, user_id),
    CONSTRAINT fk_cp_conversation
    FOREIGN KEY (conversation_id) REFERENCES conversations (id) ON DELETE CASCADE,
    CONSTRAINT fk_cp_user
    FOREIGN KEY (user_id)         REFERENCES users (id)          ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS ix_cp_conversation ON conversation_participants (conversation_id);
CREATE INDEX IF NOT EXISTS ix_cp_user         ON conversation_participants (user_id);


/* ------------------------------------------------------------
 * 3) Scoring configuration (normalized; one active row per table)
 *    Read the most recent row (or id=1) from each table at runtime.
 *    No daily caps included (per your note).
 * ----------------------------------------------------------*/

/* Core settings: threshold, optional time guard */
CREATE TABLE IF NOT EXISTS scoring_settings (
    id                      BIGSERIAL PRIMARY KEY,
    threshold               INTEGER    NOT NULL,          -- points required to unlock
    min_days_since_match    INTEGER    NOT NULL DEFAULT 0, -- optional; set 0 to ignore
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
    );

/* Text weights */
CREATE TABLE IF NOT EXISTS scoring_text (
    id                 BIGSERIAL PRIMARY KEY,
    base               NUMERIC(8,4) NOT NULL,            -- e.g., 1.00
    per_char           NUMERIC(8,4) NOT NULL,            -- e.g., 0.02
    max_per_message    NUMERIC(8,4) NOT NULL,            -- e.g., 5.00
    cooldown_seconds   INTEGER      NOT NULL,            -- e.g., 10
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

/* Voice-note weights */
CREATE TABLE IF NOT EXISTS scoring_voice (
     id                      BIGSERIAL PRIMARY KEY,
     per_second              NUMERIC(8,4) NOT NULL,       -- e.g., 0.05
    max_per_note            NUMERIC(8,4) NOT NULL,       -- e.g., 15.00
    min_duration_seconds    INTEGER      NOT NULL,       -- e.g., 3
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

/* Call weights */
CREATE TABLE IF NOT EXISTS scoring_call (
    id                      BIGSERIAL PRIMARY KEY,
    per_minute              NUMERIC(8,4) NOT NULL,       -- e.g., 3.00
    max_per_call            NUMERIC(8,4) NOT NULL,       -- e.g., 60.00
    min_duration_seconds    INTEGER      NOT NULL,       -- e.g., 30
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

/* Bonuses */
CREATE TABLE IF NOT EXISTS scoring_bonuses (
   id                      BIGSERIAL PRIMARY KEY,
   first_message_of_day    NUMERIC(8,4) NOT NULL,       -- e.g., 2.00
    reply_within_seconds    INTEGER      NOT NULL,       -- e.g., 180
    reply_bonus             NUMERIC(8,4) NOT NULL,       -- e.g., 1.00
    created_at              TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

/* Penalties / guards */
CREATE TABLE IF NOT EXISTS scoring_penalties (
     id                          BIGSERIAL PRIMARY KEY,
     duplicate_window_seconds    INTEGER      NOT NULL,   -- e.g., 3600
     max_msgs_per_minute         INTEGER      NOT NULL,   -- e.g., 4
     created_at                  TIMESTAMPTZ  NOT NULL DEFAULT now()
    );

/* Seed sensible defaults (insert only if tables are empty) */
INSERT INTO scoring_settings (threshold, min_days_since_match)
SELECT 100, 0
    WHERE NOT EXISTS (SELECT 1 FROM scoring_settings);

INSERT INTO scoring_text (base, per_char, max_per_message, cooldown_seconds)
SELECT 1.00, 0.08, 12.00, 10
    WHERE NOT EXISTS (SELECT 1 FROM scoring_text);

INSERT INTO scoring_voice (per_second, max_per_note, min_duration_seconds)
SELECT 0.05, 15.00, 3
    WHERE NOT EXISTS (SELECT 1 FROM scoring_voice);

INSERT INTO scoring_call (per_minute, max_per_call, min_duration_seconds)
SELECT 3.00, 60.00, 30
    WHERE NOT EXISTS (SELECT 1 FROM scoring_call);

INSERT INTO scoring_bonuses (first_message_of_day, reply_within_seconds, reply_bonus)
SELECT 2.00, 180, 1.00
    WHERE NOT EXISTS (SELECT 1 FROM scoring_bonuses);

INSERT INTO scoring_penalties (duplicate_window_seconds, max_msgs_per_minute)
SELECT 3600, 4
    WHERE NOT EXISTS (SELECT 1 FROM scoring_penalties);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

-- Drop scoring config tables first (no JSON anywhere)
DROP TABLE IF EXISTS scoring_penalties;
DROP TABLE IF EXISTS scoring_bonuses;
DROP TABLE IF EXISTS scoring_call;
DROP TABLE IF EXISTS scoring_voice;
DROP TABLE IF EXISTS scoring_text;
DROP TABLE IF EXISTS scoring_settings;

-- Drop participant progress
DROP INDEX IF EXISTS ix_cp_user;
DROP INDEX IF EXISTS ix_cp_conversation;
DROP TABLE IF EXISTS conversation_participants;

-- Remove columns / constraint from conversations
ALTER TABLE conversations
DROP CONSTRAINT IF EXISTS conversations_visibility_state_chk;

ALTER TABLE conversations
DROP COLUMN IF EXISTS reveal_at,
  DROP COLUMN IF EXISTS visibility_state;

-- +goose StatementEnd