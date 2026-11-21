-- +goose Up
-- Add unique constraint to session_id to prevent duplicate summaries
ALTER TABLE whiteboard_summaries ADD CONSTRAINT unique_session_summary UNIQUE (session_id);

-- +goose Down
ALTER TABLE whiteboard_summaries DROP CONSTRAINT IF EXISTS unique_session_summary;
