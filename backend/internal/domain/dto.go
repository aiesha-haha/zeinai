package domain

// RegisterRequest defines the payload for user registration
type RegisterRequest struct {
	Phone     string `json:"phone,omitempty" validate:"omitempty,min=10,max=16"`
	Email     string `json:"email,omitempty" validate:"omitempty,email"`
	IIN       string `json:"iin,omitempty" validate:"omitempty,len=12"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,safe_name,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,safe_name,min=2,max=100"`
	BirthYear int    `json:"birth_year" validate:"required"`
	Gender    string `json:"gender" validate:"required,oneof=male female"`
	Role       string `json:"role" validate:"required,oneof=student teacher school admin"`
	SchoolID   *int64 `json:"school_id,omitempty"`   // Optional for some roles
	InviteCode string `json:"invite_code,omitempty"` // For atomic join during registration
}

// LoginRequest defines the payload for user login
type LoginRequest struct {
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=10,max=16"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password" validate:"required"`
}

// StudentLoginRequest defines payload for student specific login
type StudentLoginRequest struct {
	IIN      string `json:"iin" validate:"required,len=12"`
	Password string `json:"password" validate:"required"`
}

// StudentResetInitRequest payload for starting password reset
type StudentResetInitRequest struct {
	IIN string `json:"iin" validate:"required,len=12"`
}

// StudentResetInitResponse response containing masked phone
type StudentResetInitResponse struct {
	MaskedPhone string `json:"masked_phone"`
}

// StudentResetConfirmRequest payload for confirming password reset
type StudentResetConfirmRequest struct {
	IIN         string `json:"iin" validate:"required,len=12"`
	Code        string `json:"code" validate:"required,len=4,numeric"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// LoginResponse defines the response after successful login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type AuditLogsResponse struct {
	Logs []AuditLog `json:"logs"`
}

// GoogleRegisterCompleteRequest defines the payload for completing Google OAuth registration
type GoogleRegisterCompleteRequest struct {
	GoogleID   string `json:"google_id" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	FirstName  string `json:"first_name" validate:"required,safe_name,min=2,max=100"`
	LastName   string `json:"last_name" validate:"required,safe_name,min=2,max=100"`
	IIN        string `json:"iin,omitempty" validate:"omitempty,len=12"`
	BirthYear  int    `json:"birth_year" validate:"required,min=1950,max=2015"`
	Gender     string `json:"gender" validate:"required,oneof=male female"`
	SchoolName string `json:"school_name,omitempty"`
	Picture    string `json:"picture,omitempty"`
	// Role chosen on the profile-completion screen. Only self-service roles are allowed.
	Role string `json:"role,omitempty" validate:"omitempty,oneof=student teacher"`
}
