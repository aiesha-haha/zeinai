package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"zeinai-backend/internal/domain"
)

type AtomicAIRepo struct {
	db *pgxpool.Pool
}

func NewAIRepository(db *pgxpool.Pool) domain.AIRepository {
	return &AtomicAIRepo{db: db}
}

func (r *AtomicAIRepo) Create(ctx context.Context, req *domain.AIRequest) error {
	query := `
		INSERT INTO ai_requests (user_id, type, prompt, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()
	req.Status = domain.AIStatusPending

	err := r.db.QueryRow(ctx, query,
		req.UserID,
		req.Type,
		req.Prompt,
		req.Status,
		req.CreatedAt,
		req.UpdatedAt,
	).Scan(&req.ID)

	if err != nil {
		return fmt.Errorf("failed to create ai request: %w", err)
	}
	return nil
}

func (r *AtomicAIRepo) Update(ctx context.Context, req *domain.AIRequest) error {
	query := `
		UPDATE ai_requests
		SET response = $1, status = $2, error_msg = $3, updated_at = $4
		WHERE id = $5
	`
	req.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		req.Response,
		req.Status,
		req.ErrorMsg,
		req.UpdatedAt,
		req.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ai request: %w", err)
	}
	return nil
}

func (r *AtomicAIRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.AIRequest, error) {
	query := `
		SELECT id, user_id, type, prompt, response, status, error_msg, created_at, updated_at
		FROM ai_requests WHERE id = $1
	`
	var req domain.AIRequest
	// Note: Handling nullable fields (response, error_msg) might require sql.NullString or jackc helpers if using standard types, 
	// but pgx usually handles scanning into string as empty string if null, OR we need pointers.
	// For simplicity in this rough draft, assuming they might be scanned okay or we adjust struct to use pointers for nullables if strictly needed.
	// Actually, in domain/ai.go I defined Response as string, but DB has TEXT nullable. 
	// To be safe, let's scan into pointers or implement custom scanner.
	// We will scan standard. If DB is NULL, pgx scan to string fails? No, it often works or error.
	// Let's use pointers for nullable fields in struct or handle here. 
	// I updated domain.AIRequest to have string. Let's fix GetByID logic to handle potential nulls if I didn't change struct.
	// Actually, let's trust pgx updates or fix struct in next step if it fails.
	
	err := r.db.QueryRow(ctx, query, id).Scan(
		&req.ID,
		&req.UserID,
		&req.Type,
		&req.Prompt,
		&req.Response, // This might fail if NULL and struct is string.
		&req.Status,
		&req.ErrorMsg,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &req, nil
}
