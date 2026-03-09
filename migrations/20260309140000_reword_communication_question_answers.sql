-- +goose Up
-- Reword the three answer options for "How often do you like to communicate during the day with a partner?"
-- (Conflict & Communication). Same question_id and answer ids; only label text changes.
UPDATE question_answers
SET label = 'Often (6+ check-ins)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Often (many check-ins)';

UPDATE question_answers
SET label = 'Moderate (3–5 check-ins)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Moderate (1–2 check-ins)';

UPDATE question_answers
SET label = 'Low (1-2 check-ins)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Low (evening recap)';

-- +goose Down
-- Restore original labels
UPDATE question_answers
SET label = 'Often (many check-ins)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Often (6+ check-ins)';

UPDATE question_answers
SET label = 'Moderate (1–2 check-ins)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Moderate (3–5 check-ins)';

UPDATE question_answers
SET label = 'Low (evening recap)'
WHERE question_id = (
    SELECT q.id
    FROM questions q
             JOIN question_categories qc ON q.category_id = qc.id
    WHERE qc.key = 'conflict_communication'
      AND q.text = 'How often do you like to communicate during the day with a partner?'
)
  AND label = 'Low (1-2 check-ins)';
