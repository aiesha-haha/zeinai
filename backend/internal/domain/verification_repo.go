package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// VerificationRepository handles verification code operations
type VerificationRepository interface {
	// Create creates a new verification code
	Create(ctx context.Context, code *VerificationCode) error

	// GetByPhone gets the latest active verification code for a phone
	GetByPhone(ctx context.Context, phone string, codeType VerificationType) (*VerificationCode, error)

	// GetByEmail gets the latest active verification code for an email
	GetByEmail(ctx context.Context, email string, codeType VerificationType) (*VerificationCode, error)

	// Verify marks a verification code as verified
	Verify(ctx context.Context, id uuid.UUID) error

	// IncrementAttempts increments failed verification attempts
	IncrementAttempts(ctx context.Context, id uuid.UUID) error

	// DeleteExpired deletes expired verification codes
	DeleteExpired(ctx context.Context) error

	// GetRecentCount gets count of codes sent to phone in last N minutes
	GetRecentCount(ctx context.Context, phone string, minutes int) (int, error)

	// GetRecentCountEmail gets count of codes sent to email in last N minutes
	GetRecentCountEmail(ctx context.Context, email string, minutes int) (int, error)

	// GetBlock returns blocking information for an identifier
	GetBlock(ctx context.Context, identifier string) (*VerificationBlock, error)

	// CreateBlock creates a new block for an identifier
	CreateBlock(ctx context.Context, block *VerificationBlock) error

	// DeleteBlock removes a block for an identifier
	DeleteBlock(ctx context.Context, identifier string) error

	// GetIdentifierStats returns count of codes sent to an identifier in various intervals
	GetIdentifierStats(ctx context.Context, identifier string, intervalMinutes int) (int, error)

	// GetLastCodeTime returns the creation time of the most recent code for an identifier
	GetLastCodeTime(ctx context.Context, identifier string) (time.Time, error)

	// GetAllBlocks returns all blocks for admin view
	GetAllBlocks(ctx context.Context) ([]*VerificationBlock, error)

	// DeleteRecentCodes deletes recently sent codes for an identifier
	DeleteRecentCodes(ctx context.Context, identifier string) error
}
