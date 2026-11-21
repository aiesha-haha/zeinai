-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE presentation_slides ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE presentations ADD COLUMN IF NOT EXISTS style VARCHAR(50);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
ALTER TABLE presentation_slides DROP COLUMN IF EXISTS image_url;
ALTER TABLE presentations DROP COLUMN IF EXISTS style;
