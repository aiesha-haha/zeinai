package domain

import (
	"time"

	"github.com/google/uuid"
)

// TugOfWarStatus constants
const (
	TugOfWarStatusDraft      = "draft"
	TugOfWarStatusGenerating = "generating"
	TugOfWarStatusCompleted  = "completed"
	TugOfWarStatusFailed     = "failed"
)

// TugOfWarMaterial represents a template for a Tug of War game
type TugOfWarMaterial struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	OwnerID    uuid.UUID  `json:"owner_id" db:"owner_id"`
	SchoolID   *uuid.UUID `json:"school_id,omitempty" db:"school_id"`
	Subject    string     `json:"subject" db:"subject"`
	Grade      int        `json:"grade" db:"grade"`
	Topic      string     `json:"topic" db:"topic"`
	Language   string     `json:"language" db:"language"`
	Group1Name string     `json:"group_1_name" db:"group_1_name"`
	Group2Name string     `json:"group_2_name" db:"group_2_name"`

	Status       string  `json:"status" db:"status"`
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// TugOfWarQuestion represents an AI-generated question for Tug of War
type TugOfWarQuestion struct {
	ID         uuid.UUID `json:"id" db:"id"`
	MaterialID uuid.UUID `json:"material_id" db:"material_id"`
	GroupIndex int       `json:"group_index" db:"group_index"` // 0 or 1
	Question   string    `json:"question" db:"question"`
	Options    []string  `json:"options" db:"options"` // Will be mapped to/from JSONB
	Answer     string    `json:"answer" db:"answer"`
	Order      int       `json:"order_index" db:"order_index"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// TugOfWarWithQuestions combines material metadata with its questions
type TugOfWarWithQuestions struct {
	Material  TugOfWarMaterial   `json:"material"`
	Questions []TugOfWarQuestion `json:"questions"`
}
