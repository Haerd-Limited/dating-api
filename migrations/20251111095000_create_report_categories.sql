-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS report_categories (
    id SMALLSERIAL PRIMARY KEY,
    key TEXT UNIQUE NOT NULL,
    label TEXT NOT NULL,
    sort_order SMALLINT NOT NULL DEFAULT 0
);

INSERT INTO report_categories (key, label, sort_order)
VALUES
    ('harassment', 'Harassment or bullying', 10),
    ('spam', 'Spam or scam', 20),
    ('inappropriate_content', 'Inappropriate content', 30),
    ('safety_concern', 'Safety concern', 40),
    ('other', 'Other', 50)
ON CONFLICT (key) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS report_categories;

-- +goose StatementEnd

