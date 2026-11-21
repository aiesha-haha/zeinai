package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ArticleRepository interface {
	Create(ctx context.Context, a *domain.Article) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.Article, int, error)
	Update(ctx context.Context, a *domain.Article) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ArticleStatus, errMsg *string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	CheckOwnership(ctx context.Context, id, userID uuid.UUID) (bool, error)
}

type articleRepository struct {
	db *pgxpool.Pool
}

func NewArticleRepository(db *pgxpool.Pool) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, a *domain.Article) error {
	query := `
		INSERT INTO articles (
			id, owner_id, school_id, subject, grade, topic,
			language, description, style, complexity, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		ctx, query,
		a.ID, a.OwnerID, a.SchoolID, a.Subject, a.Grade, a.Topic,
		a.Language, a.Description, a.Style, a.Complexity, a.Status,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
}

func (r *articleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error) {
	var a domain.Article
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, description, style, complexity, title, content, status, error_message,
		       created_at, updated_at, deleted_at
		FROM articles
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.OwnerID, &a.SchoolID, &a.Subject, &a.Grade, &a.Topic,
		&a.Language, &a.Description, &a.Style, &a.Complexity, &a.Title, &a.Content, &a.Status, &a.ErrorMessage,
		&a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("article not found")
	}
	return &a, err
}

func (r *articleRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.Article, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, description, style, complexity, title, content, status, error_message,
		       created_at, updated_at
		FROM articles
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Article
	for rows.Next() {
		var a domain.Article
		if err := rows.Scan(
			&a.ID, &a.OwnerID, &a.SchoolID, &a.Subject, &a.Grade, &a.Topic,
			&a.Language, &a.Description, &a.Style, &a.Complexity, &a.Title, &a.Content, &a.Status, &a.ErrorMessage,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, a)
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM articles WHERE owner_id = $1 AND deleted_at IS NULL`,
		ownerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	if list == nil {
		list = []domain.Article{}
	}
	return list, total, nil
}

func (r *articleRepository) Update(ctx context.Context, a *domain.Article) error {
	_, err := r.db.Exec(ctx, `
		UPDATE articles
		SET title = $1, content = $2, status = $3, error_message = $4, updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL
	`, a.Title, a.Content, a.Status, a.ErrorMessage, a.ID)
	return err
}

func (r *articleRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ArticleStatus, errMsg *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE articles
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`, status, errMsg, id)
	return err
}

func (r *articleRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE articles
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

func (r *articleRepository) CheckOwnership(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT owner_id FROM articles WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("article not found")
	}
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
