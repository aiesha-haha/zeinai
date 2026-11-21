-- +goose Up

-- KMJ: learning objective entered by the teacher and lesson duration (minutes).
-- Existing plans keep working: objective is empty, duration defaults to 45.
ALTER TABLE kmj_plans ADD COLUMN IF NOT EXISTS learning_objective TEXT NOT NULL DEFAULT '';
ALTER TABLE kmj_plans ADD COLUMN IF NOT EXISTS duration_minutes INTEGER NOT NULL DEFAULT 45;

-- AI Studio: image / video / feedback templates and text materials
-- (lab work, school news, certificates). One row per generation request.
CREATE TABLE IF NOT EXISTS ai_studio_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind VARCHAR(30) NOT NULL,             -- image | video | feedback | lab | news | certificate
    title VARCHAR(500) NOT NULL,
    input JSONB NOT NULL DEFAULT '{}'::jsonb,   -- what the user entered
    content JSONB,                              -- structured, editable result (text kinds)
    file_url TEXT,                              -- generated file served from /api/v1/uploads (image/video)
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | processing | completed | failed
    error_message TEXT,
    provider_job_id TEXT,                       -- e.g. OpenAI video job id, used while polling
    request_key VARCHAR(100),                   -- client generated key: makes "Create" idempotent
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ai_studio_owner_kind ON ai_studio_items(owner_id, kind) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_studio_request_key ON ai_studio_items(owner_id, request_key) WHERE request_key IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS ai_studio_items CASCADE;
ALTER TABLE kmj_plans DROP COLUMN IF EXISTS duration_minutes;
ALTER TABLE kmj_plans DROP COLUMN IF EXISTS learning_objective;
