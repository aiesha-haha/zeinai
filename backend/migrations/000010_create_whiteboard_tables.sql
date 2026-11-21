-- +goose Up
-- Create whiteboard_sessions table
CREATE TABLE IF NOT EXISTS whiteboard_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    class_id BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'active',
    snapshot JSONB, -- Stores the final state of the board for history/AI
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_whiteboard_sessions_teacher_id ON whiteboard_sessions(teacher_id);
CREATE INDEX idx_whiteboard_sessions_class_id ON whiteboard_sessions(class_id);
CREATE INDEX idx_whiteboard_sessions_status ON whiteboard_sessions(status);

-- Create whiteboard_summaries table
CREATE TABLE IF NOT EXISTS whiteboard_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES whiteboard_sessions(id) ON DELETE CASCADE,
    summary_text TEXT NOT NULL,
    explanation TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_whiteboard_summaries_session_id ON whiteboard_summaries(session_id);

-- +goose Down
DROP TABLE IF EXISTS whiteboard_summaries;
DROP TABLE IF EXISTS whiteboard_sessions;
