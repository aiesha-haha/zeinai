-- +goose Up
CREATE TABLE tug_of_war_materials (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES users(id),
    school_id BIGINT REFERENCES schools(id),
    subject VARCHAR(255) NOT NULL,
    grade INTEGER NOT NULL,
    topic VARCHAR(255) NOT NULL,
    language VARCHAR(10) NOT NULL,
    group_1_name VARCHAR(255) DEFAULT 'Group 1',
    group_2_name VARCHAR(255) DEFAULT 'Group 2',
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tow_owner ON tug_of_war_materials(owner_id);
CREATE INDEX idx_tow_status ON tug_of_war_materials(status);

CREATE TABLE tug_of_war_questions (
    id UUID PRIMARY KEY,
    material_id UUID NOT NULL REFERENCES tug_of_war_materials(id) ON DELETE CASCADE,
    group_index INTEGER NOT NULL, -- 0 for Team 1, 1 for Team 2
    question TEXT NOT NULL,
    options JSONB NOT NULL, -- Array of 4 strings
    answer TEXT NOT NULL,
    order_index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tow_questions_material ON tug_of_war_questions(material_id);

-- +goose Down
DROP TABLE IF EXISTS tug_of_war_questions;
DROP TABLE IF EXISTS tug_of_war_materials;
