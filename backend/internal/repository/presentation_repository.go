package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PresentationRepository defines all DB operations for presentations.
type PresentationRepository interface {
	// Presentation CRUD
	Create(ctx context.Context, p *domain.Presentation) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Presentation, error)
	GetWithSlides(ctx context.Context, id uuid.UUID) (*domain.PresentationWithSlides, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.Presentation, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PresentationStatus, errMsg *string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// Slides
	CreateSlidesBatch(ctx context.Context, slides []domain.PresentationSlide) error
	GetSlidesByPresentationID(ctx context.Context, presentationID uuid.UUID) ([]domain.PresentationSlide, error)

	// Ownership
	CheckOwnership(ctx context.Context, id, userID uuid.UUID) (bool, error)
}

type presentationRepository struct {
	db *pgxpool.Pool
}

func NewPresentationRepository(db *pgxpool.Pool) PresentationRepository {
	return &presentationRepository{db: db}
}

// ==================== Presentation ====================

func (r *presentationRepository) Create(ctx context.Context, p *domain.Presentation) error {
	query := `
		INSERT INTO presentations (
			id, owner_id, school_id, subject, grade, topic,
			language, slide_count, description, status, style
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		ctx, query,
		p.ID, p.OwnerID, p.SchoolID, p.Subject, p.Grade, p.Topic,
		p.Language, p.SlideCount, p.Description, p.Status, p.Style,
	).Scan(&p.CreatedAt, &p.UpdatedAt)
}

func (r *presentationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Presentation, error) {
	var p domain.Presentation
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, slide_count, description, status, error_message,
		       created_at, updated_at, deleted_at, style
		FROM presentations
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.OwnerID, &p.SchoolID, &p.Subject, &p.Grade, &p.Topic,
		&p.Language, &p.SlideCount, &p.Description, &p.Status, &p.ErrorMessage,
		&p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.Style,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("presentation not found")
	}
	return &p, err
}

func (r *presentationRepository) GetWithSlides(ctx context.Context, id uuid.UUID) (*domain.PresentationWithSlides, error) {
	p, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	slides, err := r.GetSlidesByPresentationID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.PresentationWithSlides{
		Presentation: *p,
		Slides:       slides,
	}, nil
}

func (r *presentationRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.Presentation, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, slide_count, description, status, error_message,
		       created_at, updated_at, style
		FROM presentations
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []domain.Presentation
	for rows.Next() {
		var p domain.Presentation
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.SchoolID, &p.Subject, &p.Grade, &p.Topic,
			&p.Language, &p.SlideCount, &p.Description, &p.Status, &p.ErrorMessage,
			&p.CreatedAt, &p.UpdatedAt, &p.Style,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM presentations WHERE owner_id = $1 AND deleted_at IS NULL`,
		ownerID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	if list == nil {
		list = []domain.Presentation{}
	}
	return list, total, nil
}

func (r *presentationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.PresentationStatus, errMsg *string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE presentations
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`, status, errMsg, id)
	return err
}

func (r *presentationRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		UPDATE presentations
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`, id)
	return err
}

// ==================== Slides ====================

func (r *presentationRepository) CreateSlidesBatch(ctx context.Context, slides []domain.PresentationSlide) error {
	if len(slides) == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO presentation_slides (
			id, presentation_id, slide_number, title, content, notes, layout, image_prompt, image_url
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		ON CONFLICT (presentation_id, slide_number)
		DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			notes = EXCLUDED.notes,
			layout = EXCLUDED.layout,
			image_prompt = EXCLUDED.image_prompt,
			image_url = EXCLUDED.image_url,
			updated_at = NOW()
	`
	for _, s := range slides {
		if _, err := tx.Exec(ctx, query,
			s.ID, s.PresentationID, s.SlideNumber, s.Title, s.Content, s.Notes, s.Layout, s.ImagePrompt, s.ImageURL,
		); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *presentationRepository) GetSlidesByPresentationID(ctx context.Context, presentationID uuid.UUID) ([]domain.PresentationSlide, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, presentation_id, slide_number, title, content, notes, layout, image_prompt, image_url, created_at, updated_at
		FROM presentation_slides
		WHERE presentation_id = $1
		ORDER BY slide_number ASC
	`, presentationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slides []domain.PresentationSlide
	for rows.Next() {
		var s domain.PresentationSlide
		if err := rows.Scan(
			&s.ID, &s.PresentationID, &s.SlideNumber, &s.Title,
			&s.Content, &s.Notes, &s.Layout, &s.ImagePrompt, &s.ImageURL, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		slides = append(slides, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if slides == nil {
		slides = []domain.PresentationSlide{}
	}
	return slides, nil
}

// ==================== Ownership ====================

func (r *presentationRepository) CheckOwnership(ctx context.Context, id, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	err := r.db.QueryRow(ctx,
		`SELECT owner_id FROM presentations WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("presentation not found")
	}
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
