package domain

import (
	"time"

	"github.com/google/uuid"
)

// GroupTask statuses
const (
	GroupTaskStatusDraft      = "draft"
	GroupTaskStatusGenerating = "generating"
	GroupTaskStatusCompleted  = "completed"
	GroupTaskStatusFailed     = "failed"
)

// Session statuses
const (
	SessionStatusActive    = "active"
	SessionStatusCompleted = "completed"
)

// GroupTask represents a template for a group task activity
type GroupTask struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	OwnerID   uuid.UUID  `json:"owner_id" db:"owner_id"`
	SchoolID  *uuid.UUID `json:"school_id,omitempty" db:"school_id"`
	Subject   string     `json:"subject" db:"subject"`
	Grade     int        `json:"grade" db:"grade"`
	Topic     string     `json:"topic" db:"topic"`
	CardCount int        `json:"card_count" db:"card_count"`
	Language  string     `json:"language" db:"language"`

	Status       string  `json:"status" db:"status"`
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// GroupTaskQuestion represents an AI-generated question for a task
type GroupTaskQuestion struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TaskID     uuid.UUID `json:"task_id" db:"task_id"`
	Question   string    `json:"question" db:"question"`
	Answer     string    `json:"answer" db:"answer"`
	Difficulty string    `json:"difficulty" db:"difficulty"`
	Order      int       `json:"order" db:"order_index"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// GroupTaskSession represents a live instance of a group task being played
type GroupTaskSession struct {
	ID      uuid.UUID  `json:"id" db:"id"`
	TaskID  uuid.UUID  `json:"task_id" db:"task_id"`
	ClassID *int64     `json:"class_id,omitempty" db:"class_id"`
	Status  string     `json:"status" db:"status"`
	OwnerID uuid.UUID  `json:"owner_id" db:"owner_id"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty" db:"ended_at"`
}

// GroupTaskGroup represents a group within a session
type GroupTaskGroup struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	Name      string    `json:"name" db:"name"`
	Score     int       `json:"score" db:"score"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GroupTaskGroupStudent tracks which students are in which group for a session
type GroupTaskGroupStudent struct {
	GroupID   uuid.UUID `json:"group_id" db:"group_id"`
	StudentID uuid.UUID `json:"student_id" db:"student_id"`
}

// GroupTaskWithQuestions combines task metadata with its questions
type GroupTaskWithQuestions struct {
	Task      GroupTask           `json:"task"`
	Questions []GroupTaskQuestion `json:"questions"`
}
