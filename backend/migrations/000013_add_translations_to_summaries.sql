-- +goose Up
ALTER TABLE whiteboard_summaries ADD COLUMN IF NOT EXISTS summary_kk TEXT;
ALTER TABLE whiteboard_summaries ADD COLUMN IF NOT EXISTS summary_en TEXT;
ALTER TABLE whiteboard_summaries ADD COLUMN IF NOT EXISTS explanation_kk TEXT;
ALTER TABLE whiteboard_summaries ADD COLUMN IF NOT EXISTS explanation_en TEXT;

-- +goose Down
ALTER TABLE whiteboard_summaries DROP COLUMN IF EXISTS summary_kk;
ALTER TABLE whiteboard_summaries DROP COLUMN IF EXISTS summary_en;
ALTER TABLE whiteboard_summaries DROP COLUMN IF EXISTS explanation_kk;
ALTER TABLE whiteboard_summaries DROP COLUMN IF EXISTS explanation_en;
