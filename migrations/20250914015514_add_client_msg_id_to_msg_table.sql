-- +goose Up
-- +goose StatementBegin

-- Add column if not already present
ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS client_msg_id UUID;

-- Add uniqueness constraint: each sender_id + client_msg_id must be unique
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ux_messages_sender_clientmsg'
  ) THEN
ALTER TABLE messages
    ADD CONSTRAINT ux_messages_sender_clientmsg UNIQUE (sender_id, client_msg_id);
END IF;
END$$;

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin

-- Drop uniqueness constraint
ALTER TABLE messages
DROP CONSTRAINT IF EXISTS ux_messages_sender_clientmsg;

-- Drop column
ALTER TABLE messages
DROP COLUMN IF EXISTS client_msg_id;

-- +goose StatementEnd