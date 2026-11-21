-- +goose Up
-- +goose StatementBegin
ALTER TABLE articles 
ADD COLUMN IF NOT EXISTS style VARCHAR(50) NOT NULL DEFAULT 'academic',
ADD COLUMN IF NOT EXISTS complexity VARCHAR(50) NOT NULL DEFAULT 'medium';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE articles 
DROP COLUMN IF EXISTS style,
DROP COLUMN IF EXISTS complexity;
-- +goose StatementEnd
