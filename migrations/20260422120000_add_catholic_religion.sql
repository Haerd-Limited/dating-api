-- +goose Up
-- Add Catholic as a religion option.

INSERT INTO religions (label) VALUES
    ('Catholic')
ON CONFLICT (label) DO NOTHING;

-- +goose Down

DELETE FROM religions
WHERE label = 'Catholic';
