package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserRepository defines the methods for user persistence
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByPhoneWithDeleted(ctx context.Context, phone string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByEmailWithDeleted(ctx context.Context, email string) (*User, error)
	GetByIIN(ctx context.Context, iin string) (*User, error)
	GetByIINWithDeleted(ctx context.Context, iin string) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateByID(ctx context.Context, id uuid.UUID, user *User) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	UpdateEmailVerified(ctx context.Context, id uuid.UUID) error
	IncrementStudentsCount(ctx context.Context, id uuid.UUID) error
	DecrementStudentsCount(ctx context.Context, id uuid.UUID) error
	DecreaseStudentsCount(ctx context.Context, id uuid.UUID, amount int) error
	ListByRole(ctx context.Context, role Role, limit, offset int) ([]User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Restore(ctx context.Context, user *User) error
	UpdateSubscription(ctx context.Context, id uuid.UUID, plan SubscriptionPlan, studentsLimit int, tokensLimit int, expiresAt *time.Time) error
	UpdateAvatar(ctx context.Context, id uuid.UUID, avatarURL string) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, limit, offset int) ([]AuditLog, error)
}
