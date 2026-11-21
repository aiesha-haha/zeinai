package domain

// CreateClassRequest defines the payload for creating a new class
type CreateClassRequest struct {
	Name string `json:"name" validate:"required"`
}

// AddStudentRequest defines the payload for adding a new student to a class
type AddStudentRequest struct {
	ClassID    int64  `json:"class_id" validate:"required"`
	IIN        string `json:"iin" validate:"required,len=12"`
	Phone      string `json:"phone" validate:"required,min=10,max=16"`
	Password   string `json:"password"` // Auto-generated if empty
	FirstName  string `json:"first_name" validate:"required,safe_name,min=2,max=100"`
	LastName   string `json:"last_name" validate:"required,safe_name,min=2,max=100"`
	MiddleName string `json:"middle_name,omitempty" validate:"omitempty,safe_name,min=2,max=100"`
	Gender     string `json:"gender" validate:"required,oneof=male female"`
	BirthYear  int    `json:"birth_year" validate:"required"`
}

// UpdateClassRequest defines the payload for updating a class
type UpdateClassRequest struct {
	Name string `json:"name" validate:"required"`
}

// CreateInviteRequest defines the payload for creating a class invite
type CreateInviteRequest struct {
	UsageLimit     int `json:"usage_limit" validate:"required,min=1"`
	ExpiresInHours int `json:"expires_in_hours" validate:"required,min=1,max=168"` // Max 1 week
}

type InviteInfoResponse struct {
	ClassName   string `json:"class_name"`
	TeacherName string `json:"teacher_name"`
	ExpiresAt   string `json:"expires_at"`
}
