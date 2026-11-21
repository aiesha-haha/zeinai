-- +goose Up
-- Create group_tasks table
CREATE TABLE IF NOT EXISTS group_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    school_id UUID REFERENCES users(id),
    subject VARCHAR(255) NOT NULL,
    grade INTEGER NOT NULL,
    topic TEXT NOT NULL,
    card_count INTEGER NOT NULL DEFAULT 10,
    language VARCHAR(10) NOT NULL DEFAULT 'kaz',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_group_tasks_owner ON group_tasks(owner_id);

-- Create group_task_questions table
CREATE TABLE IF NOT EXISTS group_task_questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES group_tasks(id) ON DELETE CASCADE,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    difficulty VARCHAR(50),
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_task_questions_task ON group_task_questions(task_id);

-- Create group_task_sessions table
CREATE TABLE IF NOT EXISTS group_task_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES group_tasks(id),
    class_id BIGINT NOT NULL REFERENCES classes(id),
    owner_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_group_task_sessions_task ON group_task_sessions(task_id);
CREATE INDEX idx_group_task_sessions_class ON group_task_sessions(class_id);

-- Create group_task_groups table
CREATE TABLE IF NOT EXISTS group_task_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES group_task_sessions(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_task_groups_session ON group_task_groups(session_id);

-- Create group_task_group_students table
CREATE TABLE IF NOT EXISTS group_task_group_students (
    group_id UUID NOT NULL REFERENCES group_task_groups(id) ON DELETE CASCADE,
    student_id UUID NOT NULL REFERENCES users(id),
    PRIMARY KEY (group_id, student_id)
);

-- +goose Down
DROP TABLE IF EXISTS group_task_group_students;
DROP TABLE IF EXISTS group_task_groups;
DROP TABLE IF EXISTS group_task_sessions;
DROP TABLE IF EXISTS group_task_questions;
DROP TABLE IF EXISTS group_tasks;
