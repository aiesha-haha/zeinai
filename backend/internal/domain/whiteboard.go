package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WhiteboardStatus string

const (
	WhiteboardStatusActive    WhiteboardStatus = "active"
	WhiteboardStatusCompleted WhiteboardStatus = "completed"
)

// WhiteboardSession represents a live or past drawing session
type WhiteboardSession struct {
	ID        uuid.UUID        `json:"id" db:"id"`
	TeacherID uuid.UUID        `json:"teacher_id" db:"teacher_id"`
	ClassID   int64            `json:"class_id" db:"class_id"`
	Title     string           `json:"title" db:"title"`
	Status    WhiteboardStatus `json:"status" db:"status"`
	Snapshot  interface{}      `json:"snapshot" db:"snapshot"` // JSONB content of the board
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
	EndedAt   *time.Time       `json:"ended_at,omitempty" db:"ended_at"`
	DeletedAt *time.Time       `json:"-" db:"deleted_at"`
}

// WhiteboardSummary represents the AI-generated summary of a session
type WhiteboardSummary struct {
	ID            uuid.UUID `json:"id" db:"id"`
	SessionID     uuid.UUID `json:"session_id" db:"session_id"`
	SessionTitle  string    `json:"session_title" db:"session_title"`
	SummaryText   string    `json:"summary_text" db:"summary_text"` // RU
	SummaryKK     string    `json:"summary_kk" db:"summary_kk"`
	SummaryEN     string    `json:"summary_en" db:"summary_en"`
	Explanation   string    `json:"explanation" db:"explanation"`   // RU
	ExplanationKK string    `json:"explanation_kk" db:"explanation_kk"`
	ExplanationEN string    `json:"explanation_en" db:"explanation_en"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// WhiteboardRepository defines the interface for whiteboard data access
type WhiteboardRepository interface {
	CreateSession(ctx context.Context, session *WhiteboardSession) error
	GetSession(ctx context.Context, id uuid.UUID) (*WhiteboardSession, error)
	GetActiveSessionByTeacher(ctx context.Context, teacherID uuid.UUID) (*WhiteboardSession, error)
	GetActiveSessionByClass(ctx context.Context, classID int64) (*WhiteboardSession, error)
	UpdateSession(ctx context.Context, session *WhiteboardSession) error
	
	CreateSummary(ctx context.Context, summary *WhiteboardSummary) error
	GetSummaryBySessionID(ctx context.Context, sessionID uuid.UUID) (*WhiteboardSummary, error)
	GetSummariesByClassID(ctx context.Context, classID int64) ([]WhiteboardSummary, error)
	
	ListSessionsByClass(ctx context.Context, classID int64) ([]WhiteboardSession, error)
	ListSessionsByStudent(ctx context.Context, studentID uuid.UUID) ([]WhiteboardSession, error)
}

// WhiteboardService defines the business logic for whiteboard sessions
type WhiteboardService interface {
	StartSession(ctx context.Context, teacherID uuid.UUID, classID int64, title string) (*WhiteboardSession, error)
	EndSession(ctx context.Context, sessionID uuid.UUID, snapshot interface{}) error
	GetSession(ctx context.Context, sessionID uuid.UUID) (*WhiteboardSession, error)
	GetActiveSession(ctx context.Context, userID uuid.UUID, role Role) (*WhiteboardSession, error)
	GetHistory(ctx context.Context, classID int64) ([]WhiteboardSession, error)
	GetStudentHistory(ctx context.Context, studentID uuid.UUID) ([]WhiteboardSession, error)
	GetSummary(ctx context.Context, sessionID uuid.UUID) (*WhiteboardSummary, error)
}
