-- +goose Up
ALTER TABLE whiteboard_sessions ADD COLUMN IF NOT EXISTS title TEXT;

-- +goose Down
ALTER TABLE whiteboard_sessions DROP COLUMN title;
