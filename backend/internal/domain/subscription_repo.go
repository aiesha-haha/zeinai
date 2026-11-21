package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SubscriptionRepository handles subscription-related operations
type SubscriptionRepository interface {
	// CheckTokenBalance returns (availableTokens, error) for a user
	CheckTokenBalance(ctx context.Context, userID uuid.UUID) (int, error)
	// ConsumeTokens decrements available tokens and logs usage
	ConsumeTokens(ctx context.Context, userID uuid.UUID, amount int, actionType string) error
	
	// GetSubscription returns current subscription details for a user
	GetSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	
	// ResetTokensByDay resets tokens for users whose reset day is today
	ResetTokensByDay(ctx context.Context, day int) error
	
	// UpdatePlan changes user's subscription plan
	UpdatePlan(ctx context.Context, userID uuid.UUID, plan SubscriptionPlan, expiresAt *time.Time) error
	
	// IncrementStudentCount increases student counter
	IncrementStudentCount(ctx context.Context, userID uuid.UUID) error
	
	// DecrementStudentCount decreases student counter
	DecrementStudentCount(ctx context.Context, userID uuid.UUID) error
}
