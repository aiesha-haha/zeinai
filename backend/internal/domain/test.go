package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TestStatusDraft      = "draft"
	TestStatusGenerating = "generating"
	TestStatusCompleted  = "completed"
	TestStatusFailed     = "failed"
)

type Test struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OwnerID      uuid.UUID  `json:"owner_id" db:"owner_id"`
	SchoolID     *int64     `json:"school_id,omitempty" db:"school_id"`
	Subject      string     `json:"subject" db:"subject"`
	Grade        int        `json:"grade" db:"grade"`
	Topic        string     `json:"topic" db:"topic"`
	Language     string     `json:"language" db:"language"`
	TestType     string     `json:"test_type" db:"test_type"`
	Status       string     `json:"status" db:"status"`
	ErrorMessage *string    `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type TestQuestion struct {
	ID          uuid.UUID `json:"id" db:"id"`
	TestID      uuid.UUID `json:"test_id" db:"test_id"`
	Question    string    `json:"question" db:"question"`
	Options     []string  `json:"options" db:"options"` // mapped to JSONB array
	Answer      string    `json:"answer" db:"answer"`
	Explanation *string   `json:"explanation,omitempty" db:"explanation"`
	Order       int       `json:"order_index" db:"order_index"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type TestWithQuestions struct {
	Test      Test           `json:"test"`
	Questions []TestQuestion `json:"questions"`
	Status    string         `json:"status,omitempty"`
	Score     *int           `json:"score,omitempty"`
	MaxScore  *int           `json:"max_score,omitempty"`
	// Filled for a student only after the attempt is submitted: question id -> chosen option
	StudentAnswers map[uuid.UUID]string `json:"student_answers,omitempty"`
}

type TestAssignment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TestID     uuid.UUID  `json:"test_id" db:"test_id"`
	ClassID    int64      `json:"class_id" db:"class_id"`
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	DueDate    *time.Time `json:"due_date,omitempty" db:"due_date"`
	Status     string     `json:"status" db:"status"`
}

type TestAttempt struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	AssignmentID uuid.UUID  `json:"assignment_id" db:"assignment_id"`
	StudentID    uuid.UUID  `json:"student_id" db:"student_id"`
	Score        int        `json:"score" db:"score"`
	MaxScore     int        `json:"max_score" db:"max_score"`
	Status       string     `json:"status" db:"status"` // 'in_progress', 'completed'
	StartedAt    time.Time  `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

type TestAnswer struct {
	ID            uuid.UUID `json:"id" db:"id"`
	AttemptID     uuid.UUID `json:"attempt_id" db:"attempt_id"`
	QuestionID    uuid.UUID `json:"question_id" db:"question_id"`
	StudentAnswer string    `json:"student_answer" db:"student_answer"`
	IsCorrect     bool      `json:"is_correct" db:"is_correct"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

type TestAssignmentWithDetails struct {
	TestAssignment
	TestTitle     string `json:"test_title"`
	TestType      string `json:"test_type"`
	ClassName     string `json:"class_name"`
	StudentCount  int    `json:"student_count"`
	FinishedCount int    `json:"finished_count"`
}

type StudentTestTask struct {
	AssignmentID uuid.UUID  `json:"assignment_id"`
	TestID       uuid.UUID  `json:"test_id"`
	Topic        string     `json:"topic"`
	Subject      string     `json:"subject"`
	Grade        int        `json:"grade"`
	TestType     string     `json:"test_type"`
	Status       string     `json:"status"` // 'active' or 'completed'
	AssignedAt   time.Time  `json:"assigned_at"`
	Score        *int       `json:"score,omitempty"`
	MaxScore     *int       `json:"max_score,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

type SubmitTestRequest struct {
	Answers map[uuid.UUID]string `json:"answers"`
}

type StudentResult struct {
	StudentID   uuid.UUID  `json:"student_id"`
	StudentName string     `json:"student_name"`
	Score       int        `json:"score"`
	MaxScore    int        `json:"max_score"`
	Status      string     `json:"status"`
	CompletedAt *time.Time `json:"completed_at"`
}

type TestRepository interface {
	CreateTest(ctx context.Context, t *Test) error
	GetTestByID(ctx context.Context, id uuid.UUID) (*Test, error)
	GetTestWithQuestions(ctx context.Context, id uuid.UUID) (*TestWithQuestions, error)
	GetQuestionsByTestID(ctx context.Context, testID uuid.UUID) ([]TestQuestion, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Test, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	CheckOwnership(ctx context.Context, id uuid.UUID, userID uuid.UUID) (bool, error)

	CreateQuestionsBatch(ctx context.Context, questions []TestQuestion) error

	CreateAssignment(ctx context.Context, a *TestAssignment) error
	ListAssignmentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]TestAssignmentWithDetails, error)
	GetAssignmentResults(ctx context.Context, assignmentID uuid.UUID) ([]StudentResult, error)
	DeleteAssignment(ctx context.Context, assignmentID, teacherID uuid.UUID) error
	ListAssignmentsForStudent(ctx context.Context, studentID uuid.UUID) ([]StudentTestTask, error)
	GetAssignmentByID(ctx context.Context, id uuid.UUID) (*TestAssignment, error)
	GetAttempt(ctx context.Context, assignmentID, studentID uuid.UUID) (*TestAttempt, error)
	IsTestAssignedToStudent(ctx context.Context, testID, studentID uuid.UUID) (bool, error)
	IsAssignedToClass(ctx context.Context, testID uuid.UUID, classID int64) (bool, error)
	SubmitTest(ctx context.Context, attempt *TestAttempt, answers []TestAnswer) error
	GetAttemptAnswers(ctx context.Context, attemptID uuid.UUID) (map[uuid.UUID]string, error)
}
