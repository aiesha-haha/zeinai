package domain

// UpdateProfileRequest represents user profile update request
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,safe_name,min=2,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,safe_name,min=2,max=100"`
	BirthYear *int    `json:"birth_year,omitempty" validate:"omitempty,gte=1900,lte=2020"`
	Gender    *Gender `json:"gender,omitempty" validate:"omitempty,oneof=male female"`
	SchoolID  *int64  `json:"school_id,omitempty"`
	IIN       *string `json:"iin,omitempty" validate:"omitempty,len=12"`
}

// ProfileResponse represents user profile data
type ProfileResponse struct {
	ID           string       `json:"id"`
	Phone        string       `json:"phone"`
	Email        string       `json:"email"`
	IIN          string       `json:"iin"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	BirthYear    int          `json:"birth_year"`
	Gender       Gender       `json:"gender"`
	Role         Role         `json:"role"`
	SchoolID     *int64       `json:"school_id,omitempty"`
	AvatarURL    *string      `json:"avatar_url,omitempty"`
	Subscription Subscription `json:"subscription"`
	CreatedAt    string       `json:"created_at"`
	UpdatedAt    string       `json:"updated_at"`
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required,min=8"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=100"`
}
