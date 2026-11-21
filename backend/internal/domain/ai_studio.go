package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AI Studio item kinds
const (
	StudioKindImage       = "image"
	StudioKindVideo       = "video"
	StudioKindFeedback    = "feedback"
	StudioKindLab         = "lab"
	StudioKindNews        = "news"
	StudioKindCertificate = "certificate"
)

// AI Studio item statuses
const (
	StudioStatusPending    = "pending"
	StudioStatusProcessing = "processing"
	StudioStatusCompleted  = "completed"
	StudioStatusFailed     = "failed"
)

// Token cost per generation kind
var StudioTokenCost = map[string]int{
	StudioKindImage:       3,
	StudioKindVideo:       20,
	StudioKindFeedback:    1,
	StudioKindLab:         2,
	StudioKindNews:        2,
	StudioKindCertificate: 1,
}

// IsStudioMediaKind reports whether a kind produces a file (image / video).
func IsStudioMediaKind(kind string) bool {
	return kind == StudioKindImage || kind == StudioKindVideo
}

// StudioItem is one generation request and its result.
type StudioItem struct {
	ID            uuid.UUID       `json:"id"`
	OwnerID       uuid.UUID       `json:"owner_id"`
	Kind          string          `json:"kind"`
	Title         string          `json:"title"`
	Input         json.RawMessage `json:"input"`
	Content       json.RawMessage `json:"content,omitempty"`
	FileURL       *string         `json:"file_url,omitempty"`
	Status        string          `json:"status"`
	ErrorMessage  *string         `json:"error_message,omitempty"`
	ProviderJobID *string         `json:"-"`
	RequestKey    *string         `json:"-"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// StudioRepository persists AI Studio items.
type StudioRepository interface {
	// Create inserts the item. If an item with the same (owner, request_key)
	// already exists, nothing is inserted and created == false.
	Create(ctx context.Context, item *StudioItem) (created bool, err error)
	GetByID(ctx context.Context, id uuid.UUID) (*StudioItem, error)
	GetByRequestKey(ctx context.Context, ownerID uuid.UUID, key string) (*StudioItem, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, kind string, limit int) ([]StudioItem, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg *string) error
	SetProviderJob(ctx context.Context, id uuid.UUID, jobID string) error
	SetResult(ctx context.Context, id uuid.UUID, content json.RawMessage, fileURL *string) error
	UpdateContent(ctx context.Context, id uuid.UUID, title string, content json.RawMessage) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
}
