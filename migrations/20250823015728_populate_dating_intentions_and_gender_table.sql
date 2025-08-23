-- +goose Up
-- Seed lookup values: genders & dating intentions

INSERT INTO genders (label) VALUES
                                ('Male'),
                                ('Female')
    ON CONFLICT (label) DO NOTHING;

INSERT INTO dating_intentions (label) VALUES
                                          ('Life Partner'),
                                          ('Long-term relationship'),
                                          ('Long-term relationship, open to short-term'),
                                          ('Short-term relationship, open to long-term'),
                                          ('Short-term relationship'),
                                          ('Figuring out my dating goals')
    ON CONFLICT (label) DO NOTHING;

-- +goose Down
-- Remove only the rows we inserted (keep any user-added values)

DELETE FROM dating_intentions
WHERE label IN (
                'Life Partner',
                'Long-term relationship',
                'Long-term relationship, open to short-term',
                'Short-term relationship, open to long-term',
                'Short-term relationship',
                'Figuring out my dating goals'
    );

DELETE FROM genders
WHERE label IN (
                'Male',
                'Female'
    );