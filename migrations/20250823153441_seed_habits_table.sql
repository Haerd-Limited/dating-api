-- +goose Up
-- Seed habit options

INSERT INTO habits (label) VALUES
                               ('Yes'),
                               ('Sometimes'),
                               ('No'),
                               ('Prefer not to say')
    ON CONFLICT (label) DO NOTHING;

-- +goose Down
-- Remove only the rows we inserted (leave any others alone)

DELETE FROM habits
WHERE label IN (
                'Yes',
                'Sometimes',
                'No',
                'Prefer not to say'
    );