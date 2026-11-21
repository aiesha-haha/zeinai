package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role types
type Role string

const (
	RoleAdmin   Role = "admin"
	RoleSchool  Role = "school"
	RoleTeacher Role = "teacher"
	RoleStudent Role = "student"
)

// Gender types
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// User represents a registered user in the system
type User struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	Phone           *string    `json:"phone,omitempty" db:"phone"`
	Email           *string    `json:"email,omitempty" db:"email"`
	IIN             *string    `json:"iin,omitempty" db:"iin"`
	PasswordHash    string     `json:"-" db:"password_hash"`
	FirstName       string     `json:"first_name" db:"first_name"`
	LastName        string     `json:"last_name" db:"last_name"`
	MiddleName      *string    `json:"middle_name,omitempty" db:"middle_name"`
	BirthYear       int        `json:"birth_year" db:"birth_year"`
	Gender          Gender     `json:"gender" db:"gender"`
	Role            Role       `json:"role" db:"role"`
	SchoolID        *int64     `json:"school_id,omitempty" db:"school_id"`
	GoogleID        *string    `json:"google_id,omitempty" db:"google_id"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
	AvatarURL       *string    `json:"avatar_url,omitempty" db:"avatar_url"`

	// Subscription fields
	SubscriptionPlan      SubscriptionPlan `json:"subscription_plan" db:"subscription_plan"`
	TokensUsed            int              `json:"tokens_used" db:"tokens_used"`
	TokensLimit           int              `json:"tokens_limit" db:"tokens_limit"`
	StudentsCount         int              `json:"students_count" db:"students_count"`
	StudentsLimit         int              `json:"students_limit" db:"students_limit"`
	SubscriptionStartedAt time.Time        `json:"subscription_started_at" db:"subscription_started_at"`
	SubscriptionExpiresAt *time.Time       `json:"subscription_expires_at,omitempty" db:"subscription_expires_at"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// School represents a school entity from the Unified Database of RK
type School struct {
	ID         int64     `json:"id" db:"id"`
	NameKZ     string    `json:"name_kz" db:"name_kz"`
	NameRU     string    `json:"name_ru" db:"name_ru"`
	RegionID   int       `json:"region_id" db:"region_id"`
	DistrictID int       `json:"district_id" db:"district_id"`
	Address    string    `json:"address" db:"address"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Class represents a class created by a teacher
type Class struct {
	ID           int64      `json:"id" db:"id"`
	Name         string     `json:"name" db:"name"`
	TeacherID    uuid.UUID  `json:"teacher_id" db:"teacher_id"`
	SchoolID     *int64     `json:"school_id,omitempty" db:"school_id"`
	StudentCount int        `json:"students_count" db:"students_count"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// ClassStudent represents the many-to-many relationship between classes and students
type ClassStudent struct {
	ClassID   int64     `json:"class_id" db:"class_id"`
	StudentID uuid.UUID `json:"student_id" db:"student_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ClassInvite represents an invitation link for a class
type ClassInvite struct {
	ID         int64     `json:"id" db:"id"`
	ClassID    int64     `json:"class_id" db:"class_id"`
	Code       string    `json:"code" db:"code"`
	UsageLimit int       `json:"usage_limit" db:"usage_limit"`
	UsageCount int       `json:"usage_count" db:"usage_count"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AuditLog represents a record of a system action
type AuditLog struct {
	ID        int64       `json:"id" db:"id"`
	UserID    *uuid.UUID  `json:"user_id,omitempty" db:"user_id"`
	Role      string      `json:"role,omitempty" db:"role"`
	Action    string      `json:"action" db:"action"`
	Entity    string      `json:"entity" db:"entity"`
	EntityID  string      `json:"entity_id,omitempty" db:"entity_id"`
	Payload   interface{} `json:"payload,omitempty" db:"payload"`
	Source    string      `json:"source" db:"source"`
	RequestID string      `json:"request_id,omitempty" db:"request_id"`
	IPAddress string      `json:"ip_address,omitempty" db:"ip_address"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

type CheckIINResponse struct {
	Exists    bool   `json:"exists"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
}
