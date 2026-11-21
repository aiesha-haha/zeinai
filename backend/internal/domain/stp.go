package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// LessonMethod represents a teaching method
type LessonMethod struct {
	ID                int64     `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	Goal              string    `json:"goal" db:"goal"`
	Effectiveness     string    `json:"effectiveness" db:"effectiveness"`
	Application       string    `json:"application" db:"application"`
	IntegratedModules string    `json:"integrated_modules" db:"integrated_modules"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}

// MonthlyValue represents a national value for a specific month
type MonthlyValue struct {
	ID          int    `json:"id" db:"id"`
	MonthNumber int    `json:"month_number" db:"month_number"`
	MonthNameKZ string `json:"month_name_kz" db:"month_name_kz"`
	MonthNameRU string `json:"month_name_ru" db:"month_name_ru"`
	ValueKZ     string `json:"value_kz" db:"value_kz"`
	ValueRU     string `json:"value_ru" db:"value_ru"`
}

// ShortTermPlan represents a generated lesson plan (ҚМЖ)
type ShortTermPlan struct {
	ID                 uuid.UUID       `json:"id" db:"id"`
	UserID             uuid.UUID       `json:"user_id" db:"user_id"`
	Subject            string          `json:"subject" db:"subject"`
	Grade              int             `json:"grade" db:"grade"`
	AgeRange           string          `json:"age_range" db:"age_range"`
	Topic              string          `json:"topic" db:"topic"`
	LearningObjectives string          `json:"learning_objectives" db:"learning_objectives"`
	LessonObjectives   string          `json:"lesson_objectives" db:"lesson_objectives"`
	SelectedMethods    json.RawMessage `json:"selected_methods" db:"selected_methods" swaggertype:"string"`
	MonthlyValue       string          `json:"monthly_value" db:"monthly_value"`
	Content            string          `json:"content" db:"content"`
	AdditionalInfo     json.RawMessage `json:"additional_info" db:"additional_info" swaggertype:"string"`
	Status             AIRequestStatus `json:"status" db:"status"`
	CreatedAt          time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at" db:"updated_at"`
}

// STPRequestDTO is the payload from frontend to create a plan
type STPRequestDTO struct {
	Subject            string   `json:"subject" validate:"required"`
	Grade              int      `json:"grade" validate:"required,min=1,max=11"`
	Topic              string   `json:"topic" validate:"required"`
	LearningObjectives string   `json:"learning_objectives"`
	LessonObjectives   string   `json:"lesson_objectives"`
	Month              int      `json:"month" validate:"required,min=1,max=12"`
	MethodIDs          []int64  `json:"method_ids" validate:"max=5"`
	CustomMethods      []string `json:"custom_methods" validate:"max=5"`
}
