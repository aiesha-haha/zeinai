package domain

import (
	"time"

	"github.com/google/uuid"
)

// PresentationStatus tracks the lifecycle of a presentation generation request.
type PresentationStatus string

const (
	PresentationStatusPending    PresentationStatus = "pending"
	PresentationStatusProcessing PresentationStatus = "processing"
	PresentationStatusCompleted  PresentationStatus = "completed"
	PresentationStatusFailed     PresentationStatus = "failed"
)

// Presentation represents an AI-generated presentation (slides) created for a teacher.
type Presentation struct {
	ID      uuid.UUID  `json:"id" db:"id"`
	OwnerID uuid.UUID  `json:"owner_id" db:"owner_id"`
	// Multi-tenant support
	SchoolID *int64 `json:"school_id,omitempty" db:"school_id"`

	// Input metadata
	Subject     string `json:"subject" db:"subject"`
	Grade       int    `json:"grade" db:"grade"`
	Topic       string `json:"topic" db:"topic"`
	Language    string `json:"language" db:"language"` // kaz | rus | eng
	SlideCount  int    `json:"slide_count" db:"slide_count"`
	Description string `json:"description" db:"description"`
	Style       string `json:"style" db:"style"` // minimal | creative | professional

	// Output
	Slides []PresentationSlide `json:"slides,omitempty" db:"-"`

	// Status tracking
	Status       PresentationStatus `json:"status" db:"status"`
	ErrorMessage *string            `json:"error_message,omitempty" db:"error_message"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// PresentationSlide represents a single slide within a presentation.
type PresentationSlide struct {
	ID             uuid.UUID `json:"id" db:"id"`
	PresentationID uuid.UUID `json:"presentation_id" db:"presentation_id"`
	SlideNumber    int       `json:"slide_number" db:"slide_number"`

	// Slide content
	Title   string `json:"title" db:"title"`
	Content string `json:"content" db:"content"` // Bullet points / body text
	Notes   string `json:"notes" db:"notes"`     // Speaker notes

	// Layout hint for frontend
	Layout string `json:"layout" db:"layout"` // title_only | bullets | two_column | image_text

	// Visuals
	ImagePrompt string `json:"image_prompt,omitempty" db:"image_prompt"` // For Pollinations.ai or DALL-E
	ImageURL    string `json:"image_url,omitempty" db:"image_url"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// PresentationWithSlides is the full presentation returned to the client.
type PresentationWithSlides struct {
	Presentation Presentation        `json:"presentation"`
	Slides       []PresentationSlide `json:"slides"`
}
