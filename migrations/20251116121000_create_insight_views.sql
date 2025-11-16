-- +goose Up
-- +goose StatementBegin
-- Helper views for common aggregates. Keep logic simple and index-friendly.

-- Messages by user with simple features
CREATE OR REPLACE VIEW v_message_events AS
SELECT
    e.id,
    e.occurred_at,
    e.user_id,
    (e.props->>'type') AS message_type,
    ((e.props->>'has_voice')::boolean) AS has_voice,
    COALESCE((e.props->>'len')::int, 0) AS length_chars
FROM events e
WHERE e.name = 'conversation.message_sent';

-- Swipes funnel view
CREATE OR REPLACE VIEW v_swipe_events AS
SELECT
    e.id,
    e.occurred_at,
    e.user_id AS actor_id,
    (e.props->>'action') AS action,
    NULLIF(e.props->>'prompt_id', '')::bigint AS prompt_id
FROM events e
WHERE e.name = 'interaction.swipe_created';

-- Matches
CREATE OR REPLACE VIEW v_match_events AS
SELECT
    e.id,
    e.occurred_at,
    e.user_id,
    e.props->>'match_id' AS match_id
FROM events e
WHERE e.name = 'interaction.match_created';

-- Reveal funnel
CREATE OR REPLACE VIEW v_reveal_request_events AS
SELECT e.id, e.occurred_at, e.user_id, (e.props->>'direction') AS direction
FROM events e
WHERE e.name = 'reveal.request_created';

CREATE OR REPLACE VIEW v_reveal_decision_events AS
SELECT e.id, e.occurred_at, e.user_id, ((e.props->>'accepted')::boolean) AS accepted
FROM events e
WHERE e.name = 'reveal.decision_made';

-- Prompt popularity (likes per prompt among swipe_created where action='like')
CREATE OR REPLACE VIEW v_prompt_popularity AS
SELECT
    prompt_id,
    date_trunc('week', occurred_at) AS week_start,
    COUNT(*) FILTER (WHERE action IN ('like','superlike')) AS likes,
    COUNT(*) AS total_swipes
FROM v_swipe_events
WHERE prompt_id IS NOT NULL
GROUP BY 1,2;

-- Average response time between consecutive messages in a conversation (approx via client_msg_id timing if available later)
-- Placeholder metric at user level (uses send timestamps only):
CREATE OR REPLACE VIEW v_user_message_counts AS
SELECT
    user_id,
    date_trunc('week', occurred_at) AS week_start,
    COUNT(*) AS messages_sent,
    SUM(CASE WHEN has_voice THEN 1 ELSE 0 END) AS voice_messages_sent
FROM v_message_events
GROUP BY 1,2;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS v_user_message_counts;
DROP VIEW IF EXISTS v_prompt_popularity;
DROP VIEW IF EXISTS v_reveal_decision_events;
DROP VIEW IF EXISTS v_reveal_request_events;
DROP VIEW IF EXISTS v_match_events;
DROP VIEW IF EXISTS v_swipe_events;
DROP VIEW IF EXISTS v_message_events;
-- +goose StatementEnd


