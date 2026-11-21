-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
ALTER TABLE notifications ADD COLUMN link TEXT;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back
ALTER TABLE notifications DROP COLUMN link;
