package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"zeinai-backend/internal/domain"
)

type SubscriptionService struct {
	repo domain.SubscriptionRepository
}

func NewSubscriptionService(repo domain.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// UpgradePlan upgrades user to a new subscription plan
func (s *SubscriptionService) UpgradePlan(ctx context.Context, userID uuid.UUID, plan domain.SubscriptionPlan, durationMonths int) error {
	// Calculate expiration date
	var expiresAt *time.Time
	if plan != domain.PlanFree {
		start := time.Now()

		// If renewing/upgrading, check if current subscription is still active
		current, err := s.repo.GetSubscription(ctx, userID)
		if err == nil && current.ExpiresAt != nil && current.ExpiresAt.After(time.Now()) {
			start = *current.ExpiresAt
		}

		expiry := start.AddDate(0, durationMonths, 0)
		expiresAt = &expiry
	}

	return s.repo.UpdatePlan(ctx, userID, plan, expiresAt)
}

// ResetMonthlyTokens resets tokens for all subscriptions due for reset today
func (s *SubscriptionService) ResetMonthlyTokens(ctx context.Context) error {
	day := time.Now().Day()
	if err := s.repo.ResetTokensByDay(ctx, day); err != nil {
		return fmt.Errorf("failed to reset tokens for day %d: %w", day, err)
	}
	return nil
}
