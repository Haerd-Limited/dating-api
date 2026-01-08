-- +goose Up
-- +goose StatementBegin

-- Drop CHECK constraint temporarily
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_payload_check;

-- Create new enum with correct values (changing 'voice' to 'voicenote')
CREATE TYPE message_type_new AS ENUM ('text','voicenote','system');

-- Alter column to use new enum (safe because no 'voice' data exists)
ALTER TABLE messages ALTER COLUMN type TYPE message_type_new USING type::text::message_type_new;

-- Drop old enum and rename new one
DROP TYPE message_type;
ALTER TYPE message_type_new RENAME TO message_type;

-- Recreate CHECK constraint with updated logic for 'voicenote'
ALTER TABLE messages ADD CONSTRAINT messages_payload_check CHECK (
  (type = 'text'   AND text_body IS NOT NULL AND media_key IS NULL)
  OR (type = 'voicenote'  AND media_key IS NOT NULL)
  OR (type = 'system' AND text_body IS NOT NULL)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop CHECK constraint
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_payload_check;

-- Create old enum with original values
CREATE TYPE message_type_old AS ENUM ('text','voice','system');

-- Alter column back to old enum
ALTER TABLE messages ALTER COLUMN type TYPE message_type_old USING type::text::message_type_old;

-- Drop new enum and rename old one
DROP TYPE message_type;
ALTER TYPE message_type_old RENAME TO message_type;

-- Recreate original CHECK constraint
ALTER TABLE messages ADD CONSTRAINT messages_payload_check CHECK (
  (type = 'text'   AND text_body IS NOT NULL AND media_key IS NULL)
  OR (type = 'voice'  AND media_key IS NOT NULL)
  OR (type = 'system' AND text_body IS NOT NULL)
);

-- +goose StatementEnd

