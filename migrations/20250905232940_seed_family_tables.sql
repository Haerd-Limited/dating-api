-- +goose Up
-- +goose StatementBegin

-- Add `key` column if it doesn't exist
ALTER TABLE family_statuses
    ADD COLUMN IF NOT EXISTS key TEXT UNIQUE;

ALTER TABLE family_plans
    ADD COLUMN IF NOT EXISTS key TEXT UNIQUE;

-- Seed family_statuses
INSERT INTO family_statuses (key, label) VALUES
                                             ('dont_have_children', 'Don''t have children'),
                                             ('have_children',       'Have children'),
                                             ('prefer_not_to_say',   'Prefer not to say')
    ON CONFLICT (key) DO NOTHING;

-- Seed family_plans
INSERT INTO family_plans (key, label) VALUES
                                          ('dont_want_children', 'Don''t want children'),
                                          ('want_children',      'Want children'),
                                          ('open_to_children',   'Open to children'),
                                          ('not_sure',           'Not sure'),
                                          ('prefer_not_to_say',  'Prefer not to say')
    ON CONFLICT (key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove seeded rows
DELETE FROM family_statuses WHERE key IN ('dont_have_children','have_children','prefer_not_to_say');
DELETE FROM family_plans WHERE key IN ('dont_want_children','want_children','open_to_children','not_sure','prefer_not_to_say');

-- Drop key column
ALTER TABLE family_statuses DROP COLUMN IF EXISTS key;
ALTER TABLE family_plans DROP COLUMN IF EXISTS key;
-- +goose StatementEnd