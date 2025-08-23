-- +goose Up
-- Seed lookup values to match Hinge-style options/wording

-- Religions
INSERT INTO religions (label) VALUES
                                  ('Agnostic'),
                                  ('Atheist'),
                                  ('Buddhist'),
                                  ('Christian'),
                                  ('Hindu'),
                                  ('Jewish'),
                                  ('Muslim'),
                                  ('Sikh'),
                                  ('Spiritual'),
                                  ('Other'),
                                  ('Prefer not to say')
    ON CONFLICT (label) DO NOTHING;

-- Political beliefs
INSERT INTO political_beliefs (label) VALUES
                                          ('Liberal'),
                                          ('Moderate'),
                                          ('Conservative'),
                                          ('Not political'),
                                          ('Other'),
                                          ('Prefer not to say')
    ON CONFLICT (label) DO NOTHING;

-- +goose Down
-- Remove only the rows seeded above

DELETE FROM political_beliefs
WHERE label IN (
                'Liberal',
                'Moderate',
                'Conservative',
                'Not political',
                'Other',
                'Prefer not to say'
    );

DELETE FROM religions
WHERE label IN (
                'Agnostic',
                'Atheist',
                'Buddhist',
                'Christian',
                'Hindu',
                'Jewish',
                'Muslim',
                'Sikh',
                'Spiritual',
                'Other',
                'Prefer not to say'
    );