package domain

import "time"

// SubscriptionPlan represents user subscription tier
type SubscriptionPlan string

const (
	PlanFree  SubscriptionPlan = "free"
	PlanBasic SubscriptionPlan = "basic"
	PlanUltra SubscriptionPlan = "ultra"
)

// PlanLimits defines limits for each subscription plan
type PlanLimits struct {
	StudentsLimit int
	TokensLimit   int
}

// GetPlanLimits returns limits for a given plan
func GetPlanLimits(plan SubscriptionPlan) PlanLimits {
	switch plan {
	case PlanFree:
		return PlanLimits{
			StudentsLimit: 10,
			TokensLimit:   100,
		}
	case PlanBasic:
		return PlanLimits{
			StudentsLimit: 200,
			TokensLimit:   1000,
		}
	case PlanUltra:
		return PlanLimits{
			StudentsLimit: -1, // Unlimited
			TokensLimit:   -1, // Unlimited
		}
	default:
		return GetPlanLimits(PlanFree)
	}
}

// Subscription represents user subscription details
type Subscription struct {
	Plan              SubscriptionPlan `json:"plan"`
	TokensUsed        int              `json:"tokens_used"`
	TokensLimit       int              `json:"tokens_limit"`
	TokensRemaining   int              `json:"tokens_remaining"`
	StudentsCount     int              `json:"students_count"`
	StudentsLimit     int              `json:"students_limit"`
	StudentsRemaining int              `json:"students_remaining"`
	StartedAt         time.Time        `json:"started_at"`
	ExpiresAt         *time.Time       `json:"expires_at,omitempty"`
	LimitsResetAt     *time.Time       `json:"limits_reset_at,omitempty"`
}

// CalculateNextResetDate returns the next day when monthly limits should reset
func CalculateNextResetDate(startedAt time.Time) time.Time {
	now := time.Now()
	// Default behavior: reset on the same day as started_at
	nextReset := time.Date(now.Year(), now.Month(), startedAt.Day(), startedAt.Hour(), startedAt.Minute(), startedAt.Second(), 0, now.Location())

	// If reset date for this month already passed, move to next month
	if nextReset.Before(now) {
		nextReset = nextReset.AddDate(0, 1, 0)
	}
	return nextReset
}

// CanUseTokens checks if user can consume specified amount of tokens
func (s *Subscription) CanUseTokens(amount int) bool {
	if s.Plan == PlanUltra {
		return true // Unlimited
	}
	return s.TokensUsed+amount <= s.TokensLimit
}

// CanAddStudents checks if user can add more students
func (s *Subscription) CanAddStudents(count int) bool {
	if s.Plan == PlanUltra {
		return true // Unlimited
	}
	return s.StudentsCount+count <= s.StudentsLimit
}

// TokenUsage represents a record of token consumption
type TokenUsage struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	TokensConsumed int       `json:"tokens_consumed"`
	ActionType     string    `json:"action_type"`
	CreatedAt      time.Time `json:"created_at"`
}
