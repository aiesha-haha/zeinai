package domain

// UpgradeSubscriptionRequest represents subscription upgrade request
type UpgradeSubscriptionRequest struct {
	Plan          SubscriptionPlan `json:"plan" validate:"required,oneof=free basic ultra"`
	DurationMonths int             `json:"duration_months" validate:"required,min=1,max=12"`
	PaymentMethod  string          `json:"payment_method,omitempty"` // For future payment integration
}

// SubscriptionResponse represents subscription details
type SubscriptionResponse struct {
	Plan              SubscriptionPlan `json:"plan"`
	TokensUsed        int              `json:"tokens_used"`
	TokensLimit       int              `json:"tokens_limit"`
	TokensRemaining   int              `json:"tokens_remaining"`
	StudentsCount     int              `json:"students_count"`
	StudentsLimit     int              `json:"students_limit"`
	StudentsRemaining int              `json:"students_remaining"`
	StartedAt         string           `json:"started_at"`
	ExpiresAt         *string          `json:"expires_at,omitempty"`
	LimitsResetAt     *string          `json:"limits_reset_at,omitempty"`
	IsActive          bool             `json:"is_active"`
	DaysRemaining     *int             `json:"days_remaining,omitempty"`
}
