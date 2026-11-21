package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TugOfWarRepository interface {
	CreateMaterial(ctx context.Context, material *domain.TugOfWarMaterial) error
	GetMaterialByID(ctx context.Context, id uuid.UUID) (*domain.TugOfWarMaterial, error)
	GetMaterialWithQuestions(ctx context.Context, id uuid.UUID) (*domain.TugOfWarWithQuestions, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.TugOfWarMaterial, int, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	CreateQuestionsBatch(ctx context.Context, questions []domain.TugOfWarQuestion) error
	GetQuestionsByMaterialID(ctx context.Context, materialID uuid.UUID) ([]domain.TugOfWarQuestion, error)

	CheckOwnership(ctx context.Context, materialID, userID uuid.UUID) (bool, error)
}

type tugOfWarRepository struct {
	db *pgxpool.Pool
}

func NewTugOfWarRepository(db *pgxpool.Pool) TugOfWarRepository {
	return &tugOfWarRepository{db: db}
}

func (r *tugOfWarRepository) CreateMaterial(ctx context.Context, m *domain.TugOfWarMaterial) error {
	query := `
		INSERT INTO tug_of_war_materials (
			id, owner_id, school_id, subject, grade, topic, 
			language, group_1_name, group_2_name, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		ctx, query,
		m.ID, m.OwnerID, m.SchoolID, m.Subject, m.Grade, m.Topic,
		m.Language, m.Group1Name, m.Group2Name, m.Status,
	).Scan(&m.CreatedAt, &m.UpdatedAt)
}

func (r *tugOfWarRepository) GetMaterialByID(ctx context.Context, id uuid.UUID) (*domain.TugOfWarMaterial, error) {
	var m domain.TugOfWarMaterial
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, group_1_name, group_2_name, status, error_message,
		       created_at, updated_at, deleted_at
		FROM tug_of_war_materials
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.OwnerID, &m.SchoolID, &m.Subject, &m.Grade, &m.Topic,
		&m.Language, &m.Group1Name, &m.Group2Name, &m.Status, &m.ErrorMessage,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("material not found")
	}
	return &m, err
}

func (r *tugOfWarRepository) GetMaterialWithQuestions(ctx context.Context, id uuid.UUID) (*domain.TugOfWarWithQuestions, error) {
	material, err := r.GetMaterialByID(ctx, id)
	if err != nil {
		return nil, err
	}
	questions, err := r.GetQuestionsByMaterialID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.TugOfWarWithQuestions{
		Material:  *material,
		Questions: questions,
	}, nil
}

func (r *tugOfWarRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.TugOfWarMaterial, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, group_1_name, group_2_name, status, error_message,
		       created_at, updated_at
		FROM tug_of_war_materials
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var materials []domain.TugOfWarMaterial
	for rows.Next() {
		var m domain.TugOfWarMaterial
		err := rows.Scan(
			&m.ID, &m.OwnerID, &m.SchoolID, &m.Subject, &m.Grade, &m.Topic,
			&m.Language, &m.Group1Name, &m.Group2Name, &m.Status, &m.ErrorMessage,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		materials = append(materials, m)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM tug_of_war_materials WHERE owner_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if materials == nil {
		materials = []domain.TugOfWarMaterial{}
	}

	return materials, total, nil
}

func (r *tugOfWarRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE tug_of_war_materials
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, status, errorMsg, id)
	return err
}

func (r *tugOfWarRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tug_of_war_materials
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *tugOfWarRepository) CreateQuestionsBatch(ctx context.Context, questions []domain.TugOfWarQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO tug_of_war_questions (
			id, material_id, group_index, question, options, answer, order_index
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, q := range questions {
		_, err := tx.Exec(ctx, query, q.ID, q.MaterialID, q.GroupIndex, q.Question, q.Options, q.Answer, q.Order)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *tugOfWarRepository) GetQuestionsByMaterialID(ctx context.Context, materialID uuid.UUID) ([]domain.TugOfWarQuestion, error) {
	query := `
		SELECT id, material_id, group_index, question, options, answer, order_index, created_at
		FROM tug_of_war_questions
		WHERE material_id = $1
		ORDER BY group_index, order_index ASC
	`
	rows, err := r.db.Query(ctx, query, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.TugOfWarQuestion
	for rows.Next() {
		var q domain.TugOfWarQuestion
		err := rows.Scan(&q.ID, &q.MaterialID, &q.GroupIndex, &q.Question, &q.Options, &q.Answer, &q.Order, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if questions == nil {
		questions = []domain.TugOfWarQuestion{}
	}
	return questions, nil
}

func (r *tugOfWarRepository) CheckOwnership(ctx context.Context, materialID, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	query := `SELECT owner_id FROM tug_of_war_materials WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRow(ctx, query, materialID).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("material not found")
	}
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
