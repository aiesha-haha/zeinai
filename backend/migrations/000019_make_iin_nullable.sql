-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE users ALTER COLUMN iin DROP NOT NULL;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is applied
ALTER TABLE users ALTER COLUMN iin SET NOT NULL;
