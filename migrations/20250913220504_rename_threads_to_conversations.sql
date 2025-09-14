-- +goose Up
-- +goose StatementBegin

-- 1) Rename table: threads -> conversations
ALTER TABLE IF EXISTS threads RENAME TO conversations;

-- 2) Rename the pair unique index (if you created it earlier)
ALTER INDEX IF EXISTS ux_threads_pair RENAME TO ux_conversations_pair;

-- 3) If you had a helper index on users (optional in earlier versions)
ALTER INDEX IF EXISTS idx_threads_users RENAME TO idx_conversations_users;

-- 4) Drop old FK name from conversations.last_message_id (it used to be on threads)
ALTER TABLE IF EXISTS conversations
DROP CONSTRAINT IF EXISTS threads_last_message_fk;

-- 5) Rename messages.thread_id -> messages.conversation_id
--    First drop the old FK (name depends on Postgres auto-naming if you didn't name it)
ALTER TABLE IF EXISTS messages
DROP CONSTRAINT IF EXISTS messages_thread_id_fkey;

ALTER TABLE IF EXISTS messages
    RENAME COLUMN thread_id TO conversation_id;

-- 6) Recreate the FK from messages.conversation_id -> conversations(id)
ALTER TABLE IF EXISTS messages
    ADD CONSTRAINT messages_conversation_id_fkey
    FOREIGN KEY (conversation_id)
    REFERENCES conversations(id)
    ON DELETE CASCADE;

-- 7) Recreate/rename FK from conversations.last_message_id -> messages(id)
ALTER TABLE IF EXISTS conversations
    ADD CONSTRAINT conversations_last_message_fk
    FOREIGN KEY (last_message_id)
    REFERENCES messages(id)
    DEFERRABLE INITIALLY DEFERRED;

-- 8) Rename indexes that referenced thread_id in their names
ALTER INDEX IF EXISTS idx_messages_thread_order
  RENAME TO idx_messages_conversation_order;

ALTER INDEX IF EXISTS idx_messages_sender_thread
  RENAME TO idx_messages_sender_conversation;

-- +goose StatementEnd



-- +goose Down
-- +goose StatementBegin

-- Revert index names first (messages side)
ALTER INDEX IF EXISTS idx_messages_conversation_order
  RENAME TO idx_messages_thread_order;

ALTER INDEX IF EXISTS idx_messages_sender_conversation
  RENAME TO idx_messages_sender_thread;

-- Drop FK from conversations.last_message_id (new name)
ALTER TABLE IF EXISTS conversations
DROP CONSTRAINT IF EXISTS conversations_last_message_fk;

-- Drop FK on messages.conversation_id, then rename column back
ALTER TABLE IF EXISTS messages
DROP CONSTRAINT IF EXISTS messages_conversation_id_fkey;

ALTER TABLE IF EXISTS messages
    RENAME COLUMN conversation_id TO thread_id;

-- Recreate FK to threads(id) (table will be renamed back below)
-- (We can't add it yet because table is still named conversations; add after rename.)

-- Rename pair index & users helper index back
ALTER INDEX IF EXISTS ux_conversations_pair RENAME TO ux_threads_pair;
ALTER INDEX IF EXISTS idx_conversations_users RENAME TO idx_threads_users;

-- Rename table conversations -> threads
ALTER TABLE IF EXISTS conversations RENAME TO threads;

-- Recreate FK from messages.thread_id -> threads(id)
ALTER TABLE IF EXISTS messages
    ADD CONSTRAINT messages_thread_id_fkey
    FOREIGN KEY (thread_id)
    REFERENCES threads(id)
    ON DELETE CASCADE;

-- Recreate FK from threads.last_message_id -> messages(id) with original-ish name
ALTER TABLE IF EXISTS threads
    ADD CONSTRAINT threads_last_message_fk
    FOREIGN KEY (last_message_id)
    REFERENCES messages(id)
    DEFERRABLE INITIALLY DEFERRED;

-- +goose StatementEnd