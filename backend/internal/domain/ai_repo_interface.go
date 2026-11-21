package domain

import (
	"context"

	"github.com/google/uuid"
)

type AIRepository interface {
	Create(ctx context.Context, req *AIRequest) error
	Update(ctx context.Context, req *AIRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*AIRequest, error)
}
