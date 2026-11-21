-- +goose Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP WITH TIME ZONE;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS email_verified_at;
