package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type studioRepository struct {
	db *pgxpool.Pool
}

func NewStudioRepository(db *pgxpool.Pool) domain.StudioRepository {
	return &studioRepository{db: db}
}

const studioColumns = `id, owner_id, kind, title, input, content, file_url, status,
	error_message, provider_job_id, request_key, created_at, updated_at`

func scanStudioItem(row pgx.Row) (*domain.StudioItem, error) {
	var it domain.StudioItem
	var input, content []byte
	err := row.Scan(
		&it.ID, &it.OwnerID, &it.Kind, &it.Title, &input, &content, &it.FileURL, &it.Status,
		&it.ErrorMessage, &it.ProviderJobID, &it.RequestKey, &it.CreatedAt, &it.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	it.Input = json.RawMessage(input)
	if len(content) > 0 {
		it.Content = json.RawMessage(content)
	}
	return &it, nil
}

func (r *studioRepository) Create(ctx context.Context, item *domain.StudioItem) (bool, error) {
	input := []byte(item.Input)
	if len(input) == 0 {
		input = []byte("{}")
	}
	query := `
		INSERT INTO ai_studio_items (id, owner_id, kind, title, input, status, request_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (owner_id, request_key) WHERE request_key IS NOT NULL DO NOTHING
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		item.ID, item.OwnerID, item.Kind, item.Title, input, item.Status, item.RequestKey,
	).Scan(&item.CreatedAt, &item.UpdatedAt)
	if err == pgx.ErrNoRows {
		return false, nil // duplicate request_key
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *studioRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.StudioItem, error) {
	row := r.db.QueryRow(ctx, `SELECT `+studioColumns+` FROM ai_studio_items WHERE id = $1 AND deleted_at IS NULL`, id)
	it, err := scanStudioItem(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	return it, err
}

func (r *studioRepository) GetByRequestKey(ctx context.Context, ownerID uuid.UUID, key string) (*domain.StudioItem, error) {
	row := r.db.QueryRow(ctx, `SELECT `+studioColumns+` FROM ai_studio_items WHERE owner_id = $1 AND request_key = $2`, ownerID, key)
	it, err := scanStudioItem(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("item not found")
	}
	return it, err
}

func (r *studioRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, kind string, limit int) ([]domain.StudioItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if kind != "" {
		rows, err = r.db.Query(ctx, `SELECT `+studioColumns+` FROM ai_studio_items
			WHERE owner_id = $1 AND kind = $2 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $3`, ownerID, kind, limit)
	} else {
		rows, err = r.db.Query(ctx, `SELECT `+studioColumns+` FROM ai_studio_items
			WHERE owner_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $2`, ownerID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.StudioItem{}
	for rows.Next() {
		it, err := scanStudioItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *it)
	}
	return items, rows.Err()
}

func (r *studioRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errMsg *string) error {
	_, err := r.db.Exec(ctx, `UPDATE ai_studio_items SET status = $1, error_message = $2, updated_at = NOW() WHERE id = $3`, status, errMsg, id)
	return err
}

func (r *studioRepository) SetProviderJob(ctx context.Context, id uuid.UUID, jobID string) error {
	_, err := r.db.Exec(ctx, `UPDATE ai_studio_items SET provider_job_id = $1, updated_at = NOW() WHERE id = $2`, jobID, id)
	return err
}

func (r *studioRepository) SetResult(ctx context.Context, id uuid.UUID, content json.RawMessage, fileURL *string) error {
	var c []byte
	if len(content) > 0 {
		c = []byte(content)
	}
	_, err := r.db.Exec(ctx, `UPDATE ai_studio_items
		SET content = $1, file_url = $2, status = $3, error_message = NULL, updated_at = NOW()
		WHERE id = $4`, c, fileURL, domain.StudioStatusCompleted, id)
	return err
}

func (r *studioRepository) UpdateContent(ctx context.Context, id uuid.UUID, title string, content json.RawMessage) error {
	_, err := r.db.Exec(ctx, `UPDATE ai_studio_items SET title = $1, content = $2, updated_at = NOW() WHERE id = $3`, title, []byte(content), id)
	return err
}

func (r *studioRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE ai_studio_items SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}
