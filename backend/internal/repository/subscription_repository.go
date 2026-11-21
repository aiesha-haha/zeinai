package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"zeinai-backend/internal/domain"
)

type SubscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(db *pgxpool.Pool) domain.SubscriptionRepository {
	return &SubscriptionRepo{db: db}
}

// CheckTokenBalance returns the number of available tokens for a user.
func (r *SubscriptionRepo) CheckTokenBalance(ctx context.Context, userID uuid.UUID) (int, error) {
	var tokensLimit, tokensUsed int
	err := r.db.QueryRow(ctx, `
		SELECT tokens_limit, tokens_used
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&tokensLimit, &tokensUsed)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch token balance: %w", err)
	}
	available := tokensLimit - tokensUsed
	if available < 0 {
		available = 0
	}
	return available, nil
}

// ConsumeTokens atomically checks the balance and decrements tokens.
// Returns error with "insufficient tokens" message if balance is not enough.
func (r *SubscriptionRepo) ConsumeTokens(ctx context.Context, userID uuid.UUID, amount int, actionType string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Step 1: Check balance INSIDE the transaction to prevent race conditions
	var tokensLimit, tokensUsed int
	err = tx.QueryRow(ctx, `
		SELECT tokens_limit, tokens_used
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`, userID).Scan(&tokensLimit, &tokensUsed)
	if err != nil {
		return fmt.Errorf("failed to fetch token balance: %w", err)
	}

	available := tokensLimit - tokensUsed
	if available < amount {
		return fmt.Errorf("insufficient tokens: need %d, available %d", amount, available)
	}

	// Step 2: Deduct tokens
	_, err = tx.Exec(ctx, `
		UPDATE users 
		SET tokens_used = tokens_used + $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`, amount, userID)
	if err != nil {
		return fmt.Errorf("failed to update tokens: %w", err)
	}

	// Step 3: Log usage
	_, err = tx.Exec(ctx, `
		INSERT INTO token_usage (user_id, tokens_consumed, action_type, created_at)
		VALUES ($1, $2, $3, NOW())
	`, userID, amount, actionType)
	if err != nil {
		return fmt.Errorf("failed to log token usage: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSubscription returns current subscription details for a user
func (r *SubscriptionRepo) GetSubscription(ctx context.Context, userID uuid.UUID) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := r.db.QueryRow(ctx, `
		SELECT subscription_plan, tokens_used, tokens_limit, students_count, students_limit, subscription_started_at, subscription_expires_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(
		&sub.Plan,
		&sub.TokensUsed,
		&sub.TokensLimit,
		&sub.StudentsCount,
		&sub.StudentsLimit,
		&sub.StartedAt,
		&sub.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription: %w", err)
	}

	limits := domain.GetPlanLimits(sub.Plan)
	sub.TokensRemaining = limits.TokensLimit - sub.TokensUsed
	sub.StudentsRemaining = limits.StudentsLimit - sub.StudentsCount

	return &sub, nil
}

// ResetTokensByDay resets tokens for users whose reset day is today
func (r *SubscriptionRepo) ResetTokensByDay(ctx context.Context, day int) error {
	query := `
		UPDATE users
		SET tokens_used = 0, 
		    updated_at = NOW()
		WHERE deleted_at IS NULL
		  AND (subscription_expires_at IS NULL OR subscription_expires_at > NOW())
		  AND (
		      EXTRACT(DAY FROM subscription_started_at) = $1
		      OR (
		          $1 = EXTRACT(DAY FROM (DATE_TRUNC('month', NOW()) + INTERVAL '1 month - 1 day'))
		          AND EXTRACT(DAY FROM subscription_started_at) > $1
		      )
		  )
	`
	
	_, err := r.db.Exec(ctx, query, day)
	if err != nil {
		return fmt.Errorf("failed to reset tokens for day %d: %w", day, err)
	}
	
	return nil
}

// UpdatePlan changes user's subscription plan
func (r *SubscriptionRepo) UpdatePlan(ctx context.Context, userID uuid.UUID, plan domain.SubscriptionPlan, expiresAt *time.Time) error {
	limits := domain.GetPlanLimits(plan)
	
	query := `
		UPDATE users
		SET subscription_plan = $1,
		    tokens_used = 0,
		    tokens_limit = $2,
		    students_limit = $3,
		    subscription_started_at = NOW(),
		    subscription_expires_at = $4,
		    updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`
	
	_, err := r.db.Exec(ctx, query, plan, limits.TokensLimit, limits.StudentsLimit, expiresAt, userID)
	if err != nil {
		return fmt.Errorf("failed to update subscription plan: %w", err)
	}
	
	return nil
}

// IncrementStudentCount increases student counter
func (r *SubscriptionRepo) IncrementStudentCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET students_count = students_count + 1, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment student count: %w", err)
	}
	
	return nil
}

// DecrementStudentCount decreases student counter
func (r *SubscriptionRepo) DecrementStudentCount(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET students_count = GREATEST(students_count - 1, 0), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	
	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to decrement student count: %w", err)
	}
	
	return nil
}
