package domain

import (
	"time"

	"github.com/google/uuid"
)

type ArticleStatus string

const (
	ArticleStatusPending    ArticleStatus = "pending"
	ArticleStatusProcessing ArticleStatus = "processing"
	ArticleStatusCompleted  ArticleStatus = "completed"
	ArticleStatusFailed     ArticleStatus = "failed"
)

type Article struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	OwnerID      uuid.UUID     `json:"owner_id" db:"owner_id"`
	SchoolID     *int64        `json:"school_id,omitempty" db:"school_id"`
	Subject      string        `json:"subject" db:"subject"`
	Grade        int           `json:"grade" db:"grade"`
	Topic        string        `json:"topic" db:"topic"`
	Language     string        `json:"language" db:"language"`
	Description  string        `json:"description" db:"description"`
	Style        string        `json:"style" db:"style"`
	Complexity   string        `json:"complexity" db:"complexity"`
	Title        *string       `json:"title,omitempty" db:"title"`
	Content      *string       `json:"content,omitempty" db:"content"`
	Status       ArticleStatus `json:"status" db:"status"`
	ErrorMessage *string       `json:"error_message,omitempty" db:"error_message"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time    `json:"-" db:"deleted_at"`
}
