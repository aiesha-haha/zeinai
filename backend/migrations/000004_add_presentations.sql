-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS presentations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    school_id BIGINT REFERENCES schools(id) ON DELETE SET NULL,
    subject VARCHAR(255) NOT NULL,
    grade SMALLINT NOT NULL CHECK (grade BETWEEN 1 AND 12),
    topic TEXT NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'kaz' CHECK (language IN ('kaz', 'rus', 'eng')),
    slide_count SMALLINT NOT NULL DEFAULT 10 CHECK (slide_count BETWEEN 1 AND 50),
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_presentations_owner_id ON presentations (owner_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_presentations_created_at ON presentations (created_at DESC);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS presentation_slides (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    presentation_id UUID NOT NULL REFERENCES presentations(id) ON DELETE CASCADE,
    slide_number SMALLINT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    content TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    layout VARCHAR(30) NOT NULL DEFAULT 'bullets'
        CHECK (layout IN ('title_only', 'bullets', 'two_column', 'image_text')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (presentation_id, slide_number)
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_presentation_slides_presentation_id
    ON presentation_slides (presentation_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS presentation_slides;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS presentations;
-- +goose StatementEnd
