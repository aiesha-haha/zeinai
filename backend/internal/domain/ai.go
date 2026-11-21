package domain

import (
	"time"

	"github.com/google/uuid"
)

type AIRequestStatus string

const (
	AIStatusPending    AIRequestStatus = "pending"
	AIStatusProcessing AIRequestStatus = "processing"
	AIStatusCompleted  AIRequestStatus = "completed"
	AIStatusFailed     AIRequestStatus = "failed"
)

type AIRequestType string

const (
	AIRequestTypeText  AIRequestType = "text"
	AIRequestTypeImage AIRequestType = "image"
)

// AIRequest represents a request to the AI service
type AIRequest struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	UserID    uuid.UUID       `json:"user_id" db:"user_id"`
	Type      AIRequestType   `json:"type" db:"type"`
	Prompt    string          `json:"prompt" db:"prompt"`
	Response  *string         `json:"response,omitempty" db:"response"` // Pointer for nullable
	Status    AIRequestStatus `json:"status" db:"status"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
	ErrorMsg  *string         `json:"error,omitempty" db:"error_msg"` // Pointer for nullable
}
