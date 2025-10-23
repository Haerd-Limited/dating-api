-- +goose Up
-- Haerd Dating: OkCupid-style matching (enum-free, no DO $$)

-- 1) Core lookup tables
CREATE TABLE IF NOT EXISTS question_categories (
                                                   id          BIGSERIAL PRIMARY KEY,
                                                   key         TEXT UNIQUE NOT NULL,   -- e.g., 'faith_values'
                                                   name        TEXT NOT NULL,          -- e.g., 'Faith & Values'
                                                   created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS questions (
                                         id           BIGSERIAL PRIMARY KEY,
                                         category_id  BIGINT NOT NULL REFERENCES question_categories(id) ON DELETE RESTRICT,
    text         TEXT NOT NULL,
    type         TEXT NOT NULL DEFAULT 'structured' CHECK (type IN ('structured')),
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (category_id, text)
    );

CREATE INDEX IF NOT EXISTS idx_questions_category ON questions(category_id);

CREATE TABLE IF NOT EXISTS question_answers (
                                                id           BIGSERIAL PRIMARY KEY,
                                                question_id  BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    label        TEXT NOT NULL,
    sort         INT NOT NULL DEFAULT 0,
    UNIQUE (question_id, label)
    );

CREATE INDEX IF NOT EXISTS idx_question_answers_q ON question_answers(question_id);
CREATE INDEX IF NOT EXISTS idx_question_answers_q_sort ON question_answers(question_id, sort);

-- 2) User answers (OkCupid-style)
CREATE TABLE IF NOT EXISTS user_answers (
                                            user_id                 UUID NOT NULL,
                                            question_id             BIGINT NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id               BIGINT NOT NULL REFERENCES question_answers(id) ON DELETE RESTRICT,
    acceptable_answer_ids   BIGINT[] NOT NULL DEFAULT '{}',
    importance              TEXT NOT NULL DEFAULT 'somewhat' CHECK (importance IN ('irrelevant','a_little','somewhat','very','mandatory')),
    is_private              BOOLEAN NOT NULL DEFAULT FALSE, -- TRUE = hidden on profile, still used for matching
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, question_id)
    );

CREATE INDEX IF NOT EXISTS idx_user_answers_question ON user_answers(question_id);
CREATE INDEX IF NOT EXISTS idx_user_answers_user ON user_answers(user_id);

-- 3) Importance weights config (tunable without redeploy)
CREATE TABLE IF NOT EXISTS importance_weights (
                                                  key     TEXT PRIMARY KEY CHECK (key IN ('irrelevant','a_little','somewhat','very','mandatory')),
    weight  INT NOT NULL CHECK (weight >= 0)
    );

-- -------------------------------------------------------------------
-- Seed data
-- -------------------------------------------------------------------

-- Importance weights (suggested MVP defaults)
INSERT INTO importance_weights(key, weight) VALUES
                                                ('irrelevant', 0),
                                                ('a_little',   1),
                                                ('somewhat',   3),
                                                ('very',      10),
                                                ('mandatory', 30)
    ON CONFLICT (key) DO NOTHING;

-- Categories (10)
INSERT INTO question_categories(key, name) VALUES
                                               ('faith_values',            'Faith & Values'),
                                               ('relationship_intent',     'Relationship Intent'),
                                               ('kids_family',             'Kids & Family'),
                                               ('monogamy_boundaries',     'Monogamy & Boundaries'),
                                               ('lifestyle_cleanliness',   'Lifestyle & Cleanliness'),
                                               ('substances',              'Substances'),
                                               ('money_work',              'Money & Work'),
                                               ('politics_tolerance',      'Politics & Tolerance'),
                                               ('conflict_communication',  'Conflict & Communication'),
                                               ('time_ambition',           'Time & Ambition')
    ON CONFLICT (key) DO NOTHING;

-- 50 questions (5 per category)
WITH c AS (SELECT id, key FROM question_categories)
INSERT INTO questions (category_id, text, type)
SELECT c.id, q.text, 'structured'
FROM (VALUES
          -- Faith & Values (5)
          ('faith_values',           'How important is faith or religion in your life?'),
          ('faith_values',           'How often do you participate in faith-related activities (e.g., church, prayer, study)?'),
          ('faith_values',           'Would you date someone with very different religious beliefs from your own?'),
          ('faith_values',           'Do you prefer a partner who shares your faith practices?'),
          ('faith_values',           'Is personal integrity (keeping promises, telling the truth) a non-negotiable for you?'),

          -- Relationship Intent (5)
          ('relationship_intent',    'What relationship horizon are you seeking right now?'),
          ('relationship_intent',    'How quickly do you prefer a relationship to become exclusive?'),
          ('relationship_intent',    'How important is regular quality time (e.g., weekly date night)?'),
          ('relationship_intent',    'Would you relocate within the next 2 years for the right relationship?'),
          ('relationship_intent',    'How do you feel about long-distance relationships?'),

          -- Kids & Family (5)
          ('kids_family',            'Do you want children?'),
          ('kids_family',            'If you want children, when would you ideally like to start?'),
          ('kids_family',            'How open are you to adoption or fostering?'),
          ('kids_family',            'How important is being close with extended family?'),
          ('kids_family',            'Should parenting roles be traditional, flexible, or fully shared?'),

          -- Monogamy & Boundaries (5)
          ('monogamy_boundaries',    'What relationship structure do you prefer?'),
          ('monogamy_boundaries',    'Is flirting with others while in a relationship acceptable?'),
          ('monogamy_boundaries',    'Are you comfortable with public displays of affection (PDA)?'),
          ('monogamy_boundaries',    'Is watching adult content while in a relationship acceptable?'),
          ('monogamy_boundaries',    'How important is sexual exclusivity?'),

          -- Lifestyle & Cleanliness (5)
          ('lifestyle_cleanliness',  'How tidy/organized do you prefer your home to be?'),
          ('lifestyle_cleanliness',  'How do you feel about pets in the home?'),
          ('lifestyle_cleanliness',  'Do you prefer early mornings or late nights as your routine?'),
          ('lifestyle_cleanliness',  'How often do you cook at home vs eat out?'),
          ('lifestyle_cleanliness',  'How important is regular exercise/fitness to you?'),

          -- Substances (5)
          ('substances',             'Do you drink alcohol?'),
          ('substances',             'Do you smoke cigarettes or vape?'),
          ('substances',             'What is your view on recreational cannabis?'),
          ('substances',             'What is your view on other recreational drugs?'),
          ('substances',             'How comfortable are you with your partner drinking socially without you?'),

          -- Money & Work (5)
          ('money_work',             'How would you describe your money habits?'),
          ('money_work',             'Do you keep a monthly budget or track spending?'),
          ('money_work',             'How important is career ambition in your life?'),
          ('money_work',             'How do you feel about splitting expenses with a partner?'),
          ('money_work',             'How do you feel about one partner pausing career for family needs?'),

          -- Politics & Tolerance (5)
          ('politics_tolerance',     'How important are politics in your day-to-day life?'),
          ('politics_tolerance',     'Could you be in a relationship with someone with very different political views?'),
          ('politics_tolerance',     'How comfortable are you discussing sensitive topics (religion/politics) with a partner?'),
          ('politics_tolerance',     'How important is community service or volunteering to you?'),
          ('politics_tolerance',     'Should partners share broadly similar moral/ethical views?'),

          -- Conflict & Communication (5)
          ('conflict_communication', 'What is your default conflict style?'),
          ('conflict_communication', 'When upset, do you prefer space or immediate discussion?'),
          ('conflict_communication', 'How important is apologizing and making amends after conflict?'),
          ('conflict_communication', 'How often do you like to communicate during the day with a partner?'),
          ('conflict_communication', 'How do you feel about couples therapy or counseling if needed?'),

          -- Time & Ambition (5)
          ('time_ambition',          'How structured vs spontaneous do you prefer your week to be?'),
          ('time_ambition',          'How do you prefer to spend weekends?'),
          ('time_ambition',          'How important are personal goals/hobbies outside the relationship?'),
          ('time_ambition',          'Would you prioritize travel/adventure in the next 3 years?'),
          ('time_ambition',          'Are you comfortable with different social energy levels (introvert/extrovert) in a partner?')
     ) AS q(category_key, text)
         JOIN c ON c.key = q.category_key
    ON CONFLICT (category_id, text) DO NOTHING;

-- Seed answer options per question
-- Faith & Values
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'How important is faith or religion in your life?',
                   'How often do you participate in faith-related activities (e.g., church, prayer, study)?',
                   'Would you date someone with very different religious beliefs from your own?',
                   'Do you prefer a partner who shares your faith practices?',
                   'Is personal integrity (keeping promises, telling the truth) a non-negotiable for you?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('How important is faith or religion in your life?', 'Central to my life', 1),
                   ('How important is faith or religion in your life?', 'Important but balanced', 2),
                   ('How important is faith or religion in your life?', 'Cultural/occasional', 3),
                   ('How important is faith or religion in your life?', 'Not important', 4),

                   ('How often do you participate in faith-related activities (e.g., church, prayer, study)?', 'Weekly or more', 1),
                   ('How often do you participate in faith-related activities (e.g., church, prayer, study)?', 'Monthly', 2),
                   ('How often do you participate in faith-related activities (e.g., church, prayer, study)?', 'Few times a year', 3),
                   ('How often do you participate in faith-related activities (e.g., church, prayer, study)?', 'Rarely/Never', 4),

                   ('Would you date someone with very different religious beliefs from your own?', 'Yes', 1),
                   ('Would you date someone with very different religious beliefs from your own?', 'Maybe/depends', 2),
                   ('Would you date someone with very different religious beliefs from your own?', 'No', 3),

                   ('Do you prefer a partner who shares your faith practices?', 'Prefer same practices', 1),
                   ('Do you prefer a partner who shares your faith practices?', 'Open if values align', 2),
                   ('Do you prefer a partner who shares your faith practices?', 'No preference', 3),

                   ('Is personal integrity (keeping promises, telling the truth) a non-negotiable for you?', 'Yes, non-negotiable', 1),
                   ('Is personal integrity (keeping promises, telling the truth) a non-negotiable for you?', 'Important but context matters', 2),
                   ('Is personal integrity (keeping promises, telling the truth) a non-negotiable for you?', 'Not a strict requirement', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Relationship Intent
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'What relationship horizon are you seeking right now?',
                   'How quickly do you prefer a relationship to become exclusive?',
                   'How important is regular quality time (e.g., weekly date night)?',
                   'Would you relocate within the next 2 years for the right relationship?',
                   'How do you feel about long-distance relationships?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('What relationship horizon are you seeking right now?', 'Marriage-minded (long-term)', 1),
                   ('What relationship horizon are you seeking right now?', 'Long-term, see where it goes', 2),
                   ('What relationship horizon are you seeking right now?', 'Short-term/casual', 3),
                   ('What relationship horizon are you seeking right now?', 'Unsure', 4),

                   ('How quickly do you prefer a relationship to become exclusive?', 'Within 1–3 months', 1),
                   ('How quickly do you prefer a relationship to become exclusive?', 'Within 3–6 months', 2),
                   ('How quickly do you prefer a relationship to become exclusive?', 'After 6+ months', 3),
                   ('How quickly do you prefer a relationship to become exclusive?', 'When it feels right (no timeline)', 4),

                   ('How important is regular quality time (e.g., weekly date night)?', 'Very important', 1),
                   ('How important is regular quality time (e.g., weekly date night)?', 'Somewhat important', 2),
                   ('How important is regular quality time (e.g., weekly date night)?', 'Nice but optional', 3),
                   ('How important is regular quality time (e.g., weekly date night)?', 'Not important', 4),

                   ('Would you relocate within the next 2 years for the right relationship?', 'Yes', 1),
                   ('Would you relocate within the next 2 years for the right relationship?', 'Maybe/depends', 2),
                   ('Would you relocate within the next 2 years for the right relationship?', 'No', 3),

                   ('How do you feel about long-distance relationships?', 'Open to it', 1),
                   ('How do you feel about long-distance relationships?', 'Maybe short-term only', 2),
                   ('How do you feel about long-distance relationships?', 'Not for me', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Kids & Family
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'Do you want children?',
                   'If you want children, when would you ideally like to start?',
                   'How open are you to adoption or fostering?',
                   'How important is being close with extended family?',
                   'Should parenting roles be traditional, flexible, or fully shared?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('Do you want children?', 'Yes', 1),
                   ('Do you want children?', 'Maybe/unsure', 2),
                   ('Do you want children?', 'No', 3),

                   ('If you want children, when would you ideally like to start?', 'Within 1–2 years', 1),
                   ('If you want children, when would you ideally like to start?', 'In 3–5 years', 2),
                   ('If you want children, when would you ideally like to start?', '5+ years', 3),
                   ('If you want children, when would you ideally like to start?', 'No preference/unsure', 4),

                   ('How open are you to adoption or fostering?', 'Very open', 1),
                   ('How open are you to adoption or fostering?', 'Somewhat open', 2),
                   ('How open are you to adoption or fostering?', 'Prefer biological only', 3),

                   ('How important is being close with extended family?', 'Very important', 1),
                   ('How important is being close with extended family?', 'Somewhat important', 2),
                   ('How important is being close with extended family?', 'Not important', 3),

                   ('Should parenting roles be traditional, flexible, or fully shared?', 'Traditional roles', 1),
                   ('Should parenting roles be traditional, flexible, or fully shared?', 'Flexible by season', 2),
                   ('Should parenting roles be traditional, flexible, or fully shared?', 'Fully shared/egalitarian', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Monogamy & Boundaries
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'What relationship structure do you prefer?',
                   'Is flirting with others while in a relationship acceptable?',
                   'Are you comfortable with public displays of affection (PDA)?',
                   'Is watching adult content while in a relationship acceptable?',
                   'How important is sexual exclusivity?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('What relationship structure do you prefer?', 'Monogamy', 1),
                   ('What relationship structure do you prefer?', 'Monogamish/mostly monogamy', 2),
                   ('What relationship structure do you prefer?', 'Open/consensual non-monogamy', 3),

                   ('Is flirting with others while in a relationship acceptable?', 'Not acceptable', 1),
                   ('Is flirting with others while in a relationship acceptable?', 'Mild/harmless flirting is okay', 2),
                   ('Is flirting with others while in a relationship acceptable?', 'Acceptable', 3),

                   ('Are you comfortable with public displays of affection (PDA)?', 'Comfortable', 1),
                   ('Are you comfortable with public displays of affection (PDA)?', 'A little is okay', 2),
                   ('Are you comfortable with public displays of affection (PDA)?', 'Prefer private affection', 3),

                   ('Is watching adult content while in a relationship acceptable?', 'Acceptable', 1),
                   ('Is watching adult content while in a relationship acceptable?', 'Acceptable with boundaries', 2),
                   ('Is watching adult content while in a relationship acceptable?', 'Not acceptable', 3),

                   ('How important is sexual exclusivity?', 'Mandatory', 1),
                   ('How important is sexual exclusivity?', 'Very important', 2),
                   ('How important is sexual exclusivity?', 'Not very important', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Lifestyle & Cleanliness
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'How tidy/organized do you prefer your home to be?',
                   'How do you feel about pets in the home?',
                   'Do you prefer early mornings or late nights as your routine?',
                   'How often do you cook at home vs eat out?',
                   'How important is regular exercise/fitness to you?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('How tidy/organized do you prefer your home to be?', 'Very tidy (minimal clutter)', 1),
                   ('How tidy/organized do you prefer your home to be?', 'Moderately tidy', 2),
                   ('How tidy/organized do you prefer your home to be?', 'Casual/messy is fine', 3),

                   ('How do you feel about pets in the home?', 'Love pets (want them around)', 1),
                   ('How do you feel about pets in the home?', 'Okay with pets', 2),
                   ('How do you feel about pets in the home?', 'Prefer no pets', 3),

                   ('Do you prefer early mornings or late nights as your routine?', 'Early mornings', 1),
                   ('Do you prefer early mornings or late nights as your routine?', 'Flexible/depends', 2),
                   ('Do you prefer early mornings or late nights as your routine?', 'Late nights', 3),

                   ('How often do you cook at home vs eat out?', 'Mostly cook at home', 1),
                   ('How often do you cook at home vs eat out?', 'Balanced mix', 2),
                   ('How often do you cook at home vs eat out?', 'Mostly eat out', 3),

                   ('How important is regular exercise/fitness to you?', 'Very important', 1),
                   ('How important is regular exercise/fitness to you?', 'Somewhat important', 2),
                   ('How important is regular exercise/fitness to you?', 'Not important', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Substances
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'Do you drink alcohol?',
                   'Do you smoke cigarettes or vape?',
                   'What is your view on recreational cannabis?',
                   'What is your view on other recreational drugs?',
                   'How comfortable are you with your partner drinking socially without you?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('Do you drink alcohol?', 'No', 1),
                   ('Do you drink alcohol?', 'Occasionally/socially', 2),
                   ('Do you drink alcohol?', 'Regularly', 3),

                   ('Do you smoke cigarettes or vape?', 'No', 1),
                   ('Do you smoke cigarettes or vape?', 'Occasionally', 2),
                   ('Do you smoke cigarettes or vape?', 'Yes', 3),

                   ('What is your view on recreational cannabis?', 'Prefer not at all', 1),
                   ('What is your view on recreational cannabis?', 'Okay occasionally', 2),
                   ('What is your view on recreational cannabis?', 'Comfortable/regular', 3),

                   ('What is your view on other recreational drugs?', 'Not acceptable', 1),
                   ('What is your view on other recreational drugs?', 'Case-by-case/rare use', 2),
                   ('What is your view on other recreational drugs?', 'Acceptable', 3),

                   ('How comfortable are you with your partner drinking socially without you?', 'Comfortable', 1),
                   ('How comfortable are you with your partner drinking socially without you?', 'Prefer moderation/notice', 2),
                   ('How comfortable are you with your partner drinking socially without you?', 'Uncomfortable', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Money & Work
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'How would you describe your money habits?',
                   'Do you keep a monthly budget or track spending?',
                   'How important is career ambition in your life?',
                   'How do you feel about splitting expenses with a partner?',
                   'How do you feel about one partner pausing career for family needs?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('How would you describe your money habits?', 'Saver / cautious', 1),
                   ('How would you describe your money habits?', 'Balanced', 2),
                   ('How would you describe your money habits?', 'Spender / free-flowing', 3),

                   ('Do you keep a monthly budget or track spending?', 'Yes, consistently', 1),
                   ('Do you keep a monthly budget or track spending?', 'Sometimes', 2),
                   ('Do you keep a monthly budget or track spending?', 'No', 3),

                   ('How important is career ambition in your life?', 'Very important', 1),
                   ('How important is career ambition in your life?', 'Somewhat important', 2),
                   ('How important is career ambition in your life?', 'Not important', 3),

                   ('How do you feel about splitting expenses with a partner?', 'Roughly 50/50', 1),
                   ('How do you feel about splitting expenses with a partner?', 'Proportional to income', 2),
                   ('How do you feel about splitting expenses with a partner?', 'One person pays more/most', 3),

                   ('How do you feel about one partner pausing career for family needs?', 'Supportive/open to it', 1),
                   ('How do you feel about one partner pausing career for family needs?', 'Maybe/depends on season', 2),
                   ('How do you feel about one partner pausing career for family needs?', 'Prefer not', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Politics & Tolerance
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'How important are politics in your day-to-day life?',
                   'Could you be in a relationship with someone with very different political views?',
                   'How comfortable are you discussing sensitive topics (religion/politics) with a partner?',
                   'How important is community service or volunteering to you?',
                   'Should partners share broadly similar moral/ethical views?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('How important are politics in your day-to-day life?', 'Very important', 1),
                   ('How important are politics in your day-to-day life?', 'Somewhat important', 2),
                   ('How important are politics in your day-to-day life?', 'Not important', 3),

                   ('Could you be in a relationship with someone with very different political views?', 'Yes', 1),
                   ('Could you be in a relationship with someone with very different political views?', 'Maybe/depends', 2),
                   ('Could you be in a relationship with someone with very different political views?', 'No', 3),

                   ('How comfortable are you discussing sensitive topics (religion/politics) with a partner?', 'Very comfortable', 1),
                   ('How comfortable are you discussing sensitive topics (religion/politics) with a partner?', 'Somewhat comfortable', 2),
                   ('How comfortable are you discussing sensitive topics (religion/politics) with a partner?', 'Prefer to avoid', 3),

                   ('How important is community service or volunteering to you?', 'Very important', 1),
                   ('How important is community service or volunteering to you?', 'Somewhat important', 2),
                   ('How important is community service or volunteering to you?', 'Not important', 3),

                   ('Should partners share broadly similar moral/ethical views?', 'Yes', 1),
                   ('Should partners share broadly similar moral/ethical views?', 'Some alignment is enough', 2),
                   ('Should partners share broadly similar moral/ethical views?', 'No, differences are fine', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Conflict & Communication
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'What is your default conflict style?',
                   'When upset, do you prefer space or immediate discussion?',
                   'How important is apologizing and making amends after conflict?',
                   'How often do you like to communicate during the day with a partner?',
                   'How do you feel about couples therapy or counseling if needed?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('What is your default conflict style?', 'Direct (address it now)', 1),
                   ('What is your default conflict style?', 'Collaborative (work it out)', 2),
                   ('What is your default conflict style?', 'Avoidant (cool off first)', 3),

                   ('When upset, do you prefer space or immediate discussion?', 'Space first', 1),
                   ('When upset, do you prefer space or immediate discussion?', 'Discuss soon', 2),
                   ('When upset, do you prefer space or immediate discussion?', 'Discuss immediately', 3),

                   ('How important is apologizing and making amends after conflict?', 'Very important', 1),
                   ('How important is apologizing and making amends after conflict?', 'Somewhat important', 2),
                   ('How important is apologizing and making amends after conflict?', 'Not important', 3),

                   ('How often do you like to communicate during the day with a partner?', 'Often (many check-ins)', 1),
                   ('How often do you like to communicate during the day with a partner?', 'Moderate (1–2 check-ins)', 2),
                   ('How often do you like to communicate during the day with a partner?', 'Low (evening recap)', 3),

                   ('How do you feel about couples therapy or counseling if needed?', 'Open/positive', 1),
                   ('How do you feel about couples therapy or counseling if needed?', 'Neutral/depends', 2),
                   ('How do you feel about couples therapy or counseling if needed?', 'Prefer not', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- Time & Ambition
WITH q AS (
    SELECT id, text FROM questions
    WHERE text IN (
                   'How structured vs spontaneous do you prefer your week to be?',
                   'How do you prefer to spend weekends?',
                   'How important are personal goals/hobbies outside the relationship?',
                   'Would you prioritize travel/adventure in the next 3 years?',
                   'Are you comfortable with different social energy levels (introvert/extrovert) in a partner?'
        )
)
INSERT INTO question_answers (question_id, label, sort)
SELECT q.id, a.label, a.sort
FROM q
         JOIN (VALUES
                   ('How structured vs spontaneous do you prefer your week to be?', 'Highly structured', 1),
                   ('How structured vs spontaneous do you prefer your week to be?', 'Balanced', 2),
                   ('How structured vs spontaneous do you prefer your week to be?', 'Very spontaneous', 3),

                   ('How do you prefer to spend weekends?', 'Mostly out & social', 1),
                   ('How do you prefer to spend weekends?', 'Mix of out & in', 2),
                   ('How do you prefer to spend weekends?', 'Mostly in / low-key', 3),

                   ('How important are personal goals/hobbies outside the relationship?', 'Very important', 1),
                   ('How important are personal goals/hobbies outside the relationship?', 'Somewhat important', 2),
                   ('How important are personal goals/hobbies outside the relationship?', 'Not important', 3),

                   ('Would you prioritize travel/adventure in the next 3 years?', 'Yes', 1),
                   ('Would you prioritize travel/adventure in the next 3 years?', 'Maybe/depends', 2),
                   ('Would you prioritize travel/adventure in the next 3 years?', 'No', 3),

                   ('Are you comfortable with different social energy levels (introvert/extrovert) in a partner?', 'Yes', 1),
                   ('Are you comfortable with different social energy levels (introvert/extrovert) in a partner?', 'Maybe/depends', 2),
                   ('Are you comfortable with different social energy levels (introvert/extrovert) in a partner?', 'Prefer similar energy', 3)
) AS a(text, label, sort) ON q.text = a.text;

-- +goose Down
-- Rollback: drop in dependency order
DROP TABLE IF EXISTS user_answers;
DROP TABLE IF EXISTS question_answers;
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS question_categories;
DROP TABLE IF EXISTS importance_weights;