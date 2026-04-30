-- +goose Up
-- +goose StatementBegin
INSERT INTO prompt_types (key, label, category, is_core, core_position) VALUES
    ('core_doing_with_life', 'What I am doing with my life',                'Core', TRUE, 1),
    ('core_really_good_at',  'I am really good at',                         'Core', TRUE, 2),
    ('core_favourites',      'Favourite book, movie, show, music or food',  'Core', TRUE, 3),
    ('core_thinking_about',  'I spend a lot of time thinking about',        'Core', TRUE, 4),
    ('core_friday_night',    'On a typical Friday night',                   'Core', TRUE, 5);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM prompt_types WHERE key IN (
    'core_doing_with_life','core_really_good_at','core_favourites','core_thinking_about','core_friday_night'
);
-- +goose StatementEnd
