package domain

import (
	"time"

	"github.com/google/uuid"
)

// KMJPlan represents a complete lesson plan (Қысқа мерзімді жоспар)
type KMJPlan struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	OwnerID  uuid.UUID  `json:"owner_id" db:"owner_id"`
	SchoolID *uuid.UUID `json:"school_id,omitempty" db:"school_id"`

	// Metadata
	Subject      string     `json:"subject" db:"subject"`
	Grade        int        `json:"grade" db:"grade"`
	LessonNumber int        `json:"lesson_number" db:"lesson_number"`
	LessonTopic  string     `json:"lesson_topic" db:"lesson_topic"`
	TeacherName  string     `json:"teacher_name" db:"teacher_name"`
	LessonDate   *time.Time `json:"lesson_date,omitempty" db:"lesson_date"`

	// Teacher-provided learning objective and total lesson length in minutes
	LearningObjective string `json:"learning_objective" db:"learning_objective"`
	DurationMinutes   int    `json:"duration_minutes" db:"duration_minutes"`

	// Language
	ContentLang string `json:"content_lang" db:"content_lang"`

	// Status
	Status           string     `json:"status" db:"status"`
	GenerationTaskID *uuid.UUID `json:"generation_task_id,omitempty" db:"generation_task_id"`

	// Additional fields
	NationalValueID *uuid.UUID `json:"national_value_id,omitempty" db:"national_value_id"` // ID ценности

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// KMJCell represents a single cell in the lesson plan table
type KMJCell struct {
	ID     uuid.UUID `json:"id" db:"id"`
	PlanID uuid.UUID `json:"plan_id" db:"plan_id"`

	// Positioning
	Section string `json:"section" db:"section"`   // header, lesson_start, lesson_middle, lesson_end
	CellKey string `json:"cell_key" db:"cell_key"` // bolem, pedagog_name, topic, etc.

	// Content
	CellValue string `json:"cell_value" db:"cell_value"`

	// For lesson stages
	StageColumn *string `json:"stage_column,omitempty" db:"stage_column"` // pedagog_action, student_action, assessment, resources

	// Metadata
	IsAIGenerated bool `json:"is_ai_generated" db:"is_ai_generated"`
	IsEdited      bool `json:"is_edited" db:"is_edited"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// KMJGenerationTask tracks AI generation requests
type KMJGenerationTask struct {
	ID     uuid.UUID `json:"id" db:"id"`
	PlanID uuid.UUID `json:"plan_id" db:"plan_id"`

	// Context
	UserID   uuid.UUID  `json:"user_id" db:"user_id"`
	SchoolID *uuid.UUID `json:"school_id,omitempty" db:"school_id"`

	// AI Parameters
	Prompt   string `json:"prompt" db:"prompt"`
	Language string `json:"language" db:"language"`

	// Status
	Status       string  `json:"status" db:"status"`
	RetryCount   int     `json:"retry_count" db:"retry_count"`
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`

	// Result
	AIResponseID *uuid.UUID `json:"ai_response_id,omitempty" db:"ai_response_id"`

	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`
}

// KMJPlanWithCells combines plan with all cells for full representation
type KMJPlanWithCells struct {
	Plan          KMJPlan        `json:"plan"`
	Cells         []KMJCell      `json:"cells"`
	NationalValue *NationalValue `json:"national_value,omitempty"`
}

// Consts for KMJ sections
const (
	KMJSectionHeader       = "header"
	KMJSectionLessonStart  = "lesson_start"
	KMJSectionLessonMiddle = "lesson_middle"
	KMJSectionLessonEnd    = "lesson_end"
	KMJSectionFooter       = "footer"
)

// Consts for KMJ stage columns
const (
	KMJColumnPedagogAction   = "pedagog_action"
	KMJColumnStudentAction   = "student_action"
	KMJColumnAssessment      = "assessment"
	KMJColumnResources       = "resources"
	KMJColumnHomework        = "homework"
	KMJColumnAssessmentTotal = "assessment_total"
	KMJColumnTime            = "time"
)

// Default and allowed lesson length (minutes)
const (
	KMJDefaultDuration = 45
	KMJMinDuration     = 20
	KMJMaxDuration     = 120
)

// KMJStageMinutes splits the lesson length into start / middle / end stages.
// The three values always add up to exactly the given duration
// (45 min -> 10 / 30 / 5, the classic split used before).
func KMJStageMinutes(duration int) (start, middle, end int) {
	if duration <= 0 {
		duration = KMJDefaultDuration
	}
	start = (duration*2 + 4) / 9 // ~22%
	end = (duration + 4) / 9     // ~11%
	if start < 3 {
		start = 3
	}
	if end < 3 {
		end = 3
	}
	middle = duration - start - end
	if middle < 1 {
		middle = 1
		start = (duration - middle) / 2
		end = duration - middle - start
	}
	return start, middle, end
}

// Consts for KMJ plan status
const (
	KMJStatusDraft      = "draft"
	KMJStatusGenerating = "generating"
	KMJStatusCompleted  = "completed"
	KMJStatusFailed     = "failed"
)

// Consts for generation task status
const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
)
