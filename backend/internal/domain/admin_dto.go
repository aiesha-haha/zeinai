package domain

import "github.com/google/uuid"

type CreateTeacherRequest struct {
	FirstName string  `json:"first_name" validate:"required,safe_name,min=2,max=100"`
	LastName  string  `json:"last_name" validate:"required,safe_name,min=2,max=100"`
	Email     string  `json:"email" validate:"omitempty,email"`
	Phone     string  `json:"phone" validate:"omitempty,min=10,max=16"`
	Password  string  `json:"password" validate:"required,min=8"`
	BirthYear int     `json:"birth_year" validate:"required"`
	Gender    Gender  `json:"gender" validate:"required,oneof=male female"`
}

type AdminSetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type UpdateTeacherRequest struct {
	FirstName  *string `json:"first_name"`
	LastName   *string `json:"last_name"`
	MiddleName *string `json:"middle_name"`
	BirthYear  *int    `json:"birth_year"`
	Gender     *Gender `json:"gender"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
}

type UpdateSubscriptionRequest struct {
	Plan           SubscriptionPlan `json:"plan" validate:"required"`
	StudentsLimit  int              `json:"students_limit"`
	TokensLimit    int              `json:"tokens_limit"`
	DurationMonths int              `json:"duration_months"`
}

type AdminNotifyRequest struct {
	UserID  uuid.UUID `json:"user_id" validate:"required"`
	Title   string    `json:"title" validate:"required"`
	Message string    `json:"message" validate:"required"`
	Type    string    `json:"type" validate:"required"` // info, warning, success, error
}

type AdminBulkNotifyRequest struct {
	Title   string `json:"title" validate:"required"`
	Message string `json:"message" validate:"required"`
}
