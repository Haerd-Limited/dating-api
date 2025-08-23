-- +goose Up
-- Seed Hinge-style options

-- Education levels
INSERT INTO education_levels (label) VALUES
                                         ('Secondary school'),
                                         ('Undergrad'),
                                         ('Postgrad'),
                                         ('Prefer not to say')
    ON CONFLICT (label) DO NOTHING;

-- Ethnicities
INSERT INTO ethnicities (label) VALUES
                                    ('Black/African Descent'),
                                    ('East Asian'),
                                    ('Hispanic/Latino'),
                                    ('Middle Eastern'),
                                    ('Native American'),
                                    ('Pacific Islander'),
                                    ('South Asian'),
                                    ('Southeast Asian'),
                                    ('White/Caucasian'),
                                    ('Other'),
                                    ('Prefer not to say')
    ON CONFLICT (label) DO NOTHING;

-- +goose Down
-- Remove only the rows seeded above

DELETE FROM ethnicities
WHERE label IN (
                'Black/African Descent',
                'East Asian',
                'Hispanic/Latino',
                'Middle Eastern',
                'Native American',
                'Pacific Islander',
                'South Asian',
                'Southeast Asian',
                'White/Caucasian',
                'Other',
                'Prefer not to say'
    );

DELETE FROM education_levels
WHERE label IN (
                'Secondary school',
                'Undergrad',
                'Postgrad',
                'Prefer not to say'
    );