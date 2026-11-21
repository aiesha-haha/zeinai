-- +goose Up
ALTER TABLE group_task_sessions ALTER COLUMN class_id DROP NOT NULL;

-- +goose Down
ALTER TABLE group_task_sessions ALTER COLUMN class_id SET NOT NULL;
