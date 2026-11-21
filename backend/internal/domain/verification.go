package domain

import "time"

// VerificationType represents type of verification
type VerificationType string

const (
	VerificationTypeRegistration  VerificationType = "registration"
	VerificationTypePasswordReset VerificationType = "password_reset"
	VerificationTypeEmailChange   VerificationType = "email_change"
	VerificationTypePhoneChange   VerificationType = "phone_change"
)

// VerificationCode represents a verification code
type VerificationCode struct {
	ID         string           `json:"id"`
	Phone      string           `json:"phone,omitempty"`
	Email      string           `json:"email,omitempty"`
	Code       string           `json:"-"` // Never expose in JSON
	Type       VerificationType `json:"type"`
	IsVerified bool             `json:"is_verified"`
	Attempts   int              `json:"attempts"`
	ExpiresAt  time.Time        `json:"expires_at"`
	CreatedAt  time.Time        `json:"created_at"`
	VerifiedAt *time.Time       `json:"verified_at,omitempty"`
}

// VerificationBlock represents a blocked identifier
type VerificationBlock struct {
	ID                   string     `json:"id"`
	Identifier           string     `json:"identifier"`
	Type                 string     `json:"type"` // 'phone', 'email'
	BlockedUntil         *time.Time `json:"blocked_until,omitempty"`
	IsPermanentlyBlocked bool       `json:"is_permanently_blocked"`
	Reason               string     `json:"reason,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// VerificationStatusResponse represents the status of verification for an identifier
type VerificationStatusResponse struct {
	CanSend      bool       `json:"can_send"`
	NextSendAt   *time.Time `json:"next_send_at,omitempty"`
	Blocked      bool       `json:"blocked"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
	Reason       string     `json:"reason,omitempty"`
}

// SendVerificationCodeRequest represents request to send verification code
type SendVerificationCodeRequest struct {
	Phone string           `json:"phone,omitempty" validate:"omitempty,min=10,max=16"`
	Email string           `json:"email,omitempty" validate:"omitempty,email"`
	Type  VerificationType `json:"type" validate:"required,oneof=registration password_reset email_change phone_change"`
}

// VerifyCodeRequest represents request to verify code
type VerifyCodeRequest struct {
	Phone string `json:"phone,omitempty" validate:"omitempty,min=10,max=16"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Code  string `json:"code" validate:"required,len=6,numeric"`
}

// ResetPasswordRequest represents password reset request
type ResetPasswordRequest struct {
	Phone       string `json:"phone,omitempty" validate:"omitempty,min=10,max=16"`
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	Code        string `json:"code" validate:"required,len=6,numeric"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}
