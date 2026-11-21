-- +goose Up
-- +goose StatementBegin
ALTER TABLE presentations ADD COLUMN IF NOT EXISTS style VARCHAR(50) NOT NULL DEFAULT 'minimal';
ALTER TABLE presentation_slides ADD COLUMN IF NOT EXISTS image_prompt TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE presentation_slides DROP COLUMN IF EXISTS image_prompt;
ALTER TABLE presentations DROP COLUMN IF EXISTS style;
-- +goose StatementEnd
