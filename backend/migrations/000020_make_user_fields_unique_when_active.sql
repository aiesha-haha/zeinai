-- +goose Up
-- Drop existing unique constraints
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_phone_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_iin_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_google_id_key;

-- Create partial unique indexes
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_active_idx ON users (email) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_phone_unique_active_idx ON users (phone) WHERE deleted_at IS NULL AND phone IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_iin_unique_active_idx ON users (iin) WHERE deleted_at IS NULL AND iin IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS users_google_id_unique_active_idx ON users (google_id) WHERE deleted_at IS NULL AND google_id IS NOT NULL;

-- +goose Down
-- Drop partial unique indexes
DROP INDEX IF EXISTS users_email_unique_active_idx;
DROP INDEX IF EXISTS users_phone_unique_active_idx;
DROP INDEX IF EXISTS users_iin_unique_active_idx;
DROP INDEX IF EXISTS users_google_id_unique_active_idx;

-- Re-add standard unique constraints (this might fail if there are duplicate active/inactive records)
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE users ADD CONSTRAINT users_phone_key UNIQUE (phone);
ALTER TABLE users ADD CONSTRAINT users_iin_key UNIQUE (iin);
ALTER TABLE users ADD CONSTRAINT users_google_id_key UNIQUE (google_id);
