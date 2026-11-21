package repository

import (
	"context"
	"fmt"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VerificationRepo struct {
	db *pgxpool.Pool
}

func NewVerificationRepository(db *pgxpool.Pool) domain.VerificationRepository {
	return &VerificationRepo{db: db}
}

// Create creates a new verification code
func (r *VerificationRepo) Create(ctx context.Context, code *domain.VerificationCode) error {
	query := `
		INSERT INTO verification_codes (phone, email, code, type, is_verified, attempts, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id uuid.UUID
	err := r.db.QueryRow(ctx, query,
		code.Phone,
		code.Email,
		code.Code,
		code.Type,
		code.IsVerified,
		code.Attempts,
		code.ExpiresAt,
		code.CreatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create verification code: %w", err)
	}

	code.ID = id.String()
	return nil
}

// GetByPhone gets the latest active verification code for a phone
func (r *VerificationRepo) GetByPhone(ctx context.Context, phone string, codeType domain.VerificationType) (*domain.VerificationCode, error) {
	query := `
		SELECT id, phone, email, code, type, is_verified, attempts, expires_at, created_at, verified_at
		FROM verification_codes
		WHERE phone = $1 AND type = $2 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var code domain.VerificationCode
	var id uuid.UUID

	err := r.db.QueryRow(ctx, query, phone, codeType).Scan(
		&id,
		&code.Phone,
		&code.Email,
		&code.Code,
		&code.Type,
		&code.IsVerified,
		&code.Attempts,
		&code.ExpiresAt,
		&code.CreatedAt,
		&code.VerifiedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get verification code by phone: %w", err)
	}

	code.ID = id.String()
	return &code, nil
}

// GetByEmail gets the latest active verification code for an email
func (r *VerificationRepo) GetByEmail(ctx context.Context, email string, codeType domain.VerificationType) (*domain.VerificationCode, error) {
	query := `
		SELECT id, phone, email, code, type, is_verified, attempts, expires_at, created_at, verified_at
		FROM verification_codes
		WHERE email = $1 AND type = $2 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var code domain.VerificationCode
	var id uuid.UUID

	err := r.db.QueryRow(ctx, query, email, codeType).Scan(
		&id,
		&code.Phone,
		&code.Email,
		&code.Code,
		&code.Type,
		&code.IsVerified,
		&code.Attempts,
		&code.ExpiresAt,
		&code.CreatedAt,
		&code.VerifiedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get verification code by email: %w", err)
	}

	code.ID = id.String()
	return &code, nil
}

// Verify marks a verification code as verified
func (r *VerificationRepo) Verify(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE verification_codes
		SET is_verified = true, verified_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	return nil
}

// IncrementAttempts increments failed verification attempts
func (r *VerificationRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE verification_codes
		SET attempts = attempts + 1
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	return nil
}

// DeleteExpired deletes expired verification codes
func (r *VerificationRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM verification_codes WHERE expires_at < NOW()`

	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired codes: %w", err)
	}

	return nil
}

// GetRecentCount gets count of codes sent to phone in last N minutes
func (r *VerificationRepo) GetRecentCount(ctx context.Context, phone string, minutes int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM verification_codes
		WHERE phone = $1 AND created_at > NOW() - INTERVAL '1 minute' * $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, phone, minutes).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent count: %w", err)
	}

	return count, nil
}

// GetRecentCountEmail gets count of codes sent to email in last N minutes
func (r *VerificationRepo) GetRecentCountEmail(ctx context.Context, email string, minutes int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM verification_codes
		WHERE email = $1 AND created_at > NOW() - INTERVAL '1 minute' * $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, email, minutes).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get recent count for email: %w", err)
	}

	return count, nil
}

// GetBlock returns blocking information for an identifier
func (r *VerificationRepo) GetBlock(ctx context.Context, identifier string) (*domain.VerificationBlock, error) {
	query := `
		SELECT id, identifier, type, blocked_until, is_permanently_blocked, reason, created_at, updated_at
		FROM verification_blocks
		WHERE identifier = $1
	`

	var block domain.VerificationBlock
	var id uuid.UUID
	err := r.db.QueryRow(ctx, query, identifier).Scan(
		&id,
		&block.Identifier,
		&block.Type,
		&block.BlockedUntil,
		&block.IsPermanentlyBlocked,
		&block.Reason,
		&block.CreatedAt,
		&block.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get block: %w", err)
	}

	block.ID = id.String()
	return &block, nil
}

// CreateBlock creates a new block for an identifier
func (r *VerificationRepo) CreateBlock(ctx context.Context, block *domain.VerificationBlock) error {
	query := `
		INSERT INTO verification_blocks (identifier, type, blocked_until, is_permanently_blocked, reason)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (identifier) DO UPDATE SET
			blocked_until = EXCLUDED.blocked_until,
			is_permanently_blocked = EXCLUDED.is_permanently_blocked,
			reason = EXCLUDED.reason,
			updated_at = NOW()
	`

	_, err := r.db.Exec(ctx, query,
		block.Identifier,
		block.Type,
		block.BlockedUntil,
		block.IsPermanentlyBlocked,
		block.Reason,
	)

	if err != nil {
		return fmt.Errorf("failed to create block: %w", err)
	}

	return nil
}

// DeleteBlock removes a block for an identifier
func (r *VerificationRepo) DeleteBlock(ctx context.Context, identifier string) error {
	query := `DELETE FROM verification_blocks WHERE identifier = $1`
	_, err := r.db.Exec(ctx, query, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete block: %w", err)
	}
	return nil
}

// GetIdentifierStats returns count of codes sent to an identifier in various intervals
func (r *VerificationRepo) GetIdentifierStats(ctx context.Context, identifier string, intervalMinutes int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM verification_codes
		WHERE (phone = $1 OR email = $1) AND created_at > NOW() - INTERVAL '1 minute' * $2
	`

	var count int
	err := r.db.QueryRow(ctx, query, identifier, intervalMinutes).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get identifier stats: %w", err)
	}

	return count, nil
}

// GetLastCodeTime returns the creation time of the most recent code for an identifier
func (r *VerificationRepo) GetLastCodeTime(ctx context.Context, identifier string) (time.Time, error) {
	query := `
		SELECT created_at
		FROM verification_codes
		WHERE phone = $1 OR email = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, identifier).Scan(&createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get last code time: %w", err)
	}

	return createdAt, nil
}

// GetAllBlocks returns all blocks for admin view
func (r *VerificationRepo) GetAllBlocks(ctx context.Context) ([]*domain.VerificationBlock, error) {
	query := `
		SELECT id, identifier, type, blocked_until, is_permanently_blocked, reason, created_at, updated_at
		FROM verification_blocks
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all blocks: %w", err)
	}
	defer rows.Close()

	blocks := make([]*domain.VerificationBlock, 0)
	for rows.Next() {
		var block domain.VerificationBlock
		var id uuid.UUID
		err := rows.Scan(
			&id,
			&block.Identifier,
			&block.Type,
			&block.BlockedUntil,
			&block.IsPermanentlyBlocked,
			&block.Reason,
			&block.CreatedAt,
			&block.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan block: %w", err)
		}
		block.ID = id.String()
		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// DeleteRecentCodes deletes recently sent codes for an identifier
func (r *VerificationRepo) DeleteRecentCodes(ctx context.Context, identifier string) error {
	query := `DELETE FROM verification_codes WHERE phone = $1 OR email = $1`
	_, err := r.db.Exec(ctx, query, identifier)
	if err != nil {
		return fmt.Errorf("failed to delete recent codes: %w", err)
	}
	return nil
}
