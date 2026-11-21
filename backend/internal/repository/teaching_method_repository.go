package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeachingMethodRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TeachingMethod, error)
	List(ctx context.Context) ([]domain.TeachingMethod, error)

	// Plan methods (max 4)
	SetPlanMethods(ctx context.Context, planID uuid.UUID, methodIDs []uuid.UUID) error
	GetPlanMethods(ctx context.Context, planID uuid.UUID) ([]domain.TeachingMethod, error)
}

type teachingMethodRepository struct {
	db *pgxpool.Pool
}

func NewTeachingMethodRepository(db *pgxpool.Pool) TeachingMethodRepository {
	return &teachingMethodRepository{db: db}
}

func (r *teachingMethodRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TeachingMethod, error) {
	var method domain.TeachingMethod
	query := `
		SELECT id, name_kaz, name_rus, name_eng, type_kaz,
		       purpose_kaz, purpose_rus, purpose_eng, modules,
		       created_at, updated_at, deleted_at
		FROM teaching_methods
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&method.ID, &method.NameKaz, &method.NameRus, &method.NameEng, &method.TypeKaz,
		&method.PurposeKaz, &method.PurposeRus, &method.PurposeEng, &method.Modules,
		&method.CreatedAt, &method.UpdatedAt, &method.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("teaching method not found")
	}
	return &method, err
}

func (r *teachingMethodRepository) List(ctx context.Context) ([]domain.TeachingMethod, error) {
	query := `
		SELECT id, name_kaz, name_rus, name_eng, type_kaz,
		       purpose_kaz, purpose_rus, purpose_eng, modules,
		       created_at, updated_at, deleted_at
		FROM teaching_methods
		WHERE deleted_at IS NULL
		ORDER BY name_kaz ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []domain.TeachingMethod
	for rows.Next() {
		var method domain.TeachingMethod
		err := rows.Scan(
			&method.ID, &method.NameKaz, &method.NameRus, &method.NameEng, &method.TypeKaz,
			&method.PurposeKaz, &method.PurposeRus, &method.PurposeEng, &method.Modules,
			&method.CreatedAt, &method.UpdatedAt, &method.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	if methods == nil {
		methods = []domain.TeachingMethod{}
	}

	return methods, rows.Err()
}

func (r *teachingMethodRepository) SetPlanMethods(ctx context.Context, planID uuid.UUID, methodIDs []uuid.UUID) error {
	if len(methodIDs) > 4 {
		return fmt.Errorf("maximum 4 methods allowed, got %d", len(methodIDs))
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete existing methods for this plan
	_, err = tx.Exec(ctx, "DELETE FROM kmj_plan_methods WHERE plan_id = $1", planID)
	if err != nil {
		return fmt.Errorf("failed to delete existing methods: %w", err)
	}

	// Insert new methods
	for i, methodID := range methodIDs {
		query := `
			INSERT INTO kmj_plan_methods (id, plan_id, method_id, sort_order)
			VALUES ($1, $2, $3, $4)
		`
		_, err = tx.Exec(ctx, query, uuid.New(), planID, methodID, i+1)
		if err != nil {
			return fmt.Errorf("failed to insert method: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *teachingMethodRepository) GetPlanMethods(ctx context.Context, planID uuid.UUID) ([]domain.TeachingMethod, error) {
	query := `
		SELECT tm.id, tm.name_kaz, tm.name_rus, tm.name_eng, tm.type_kaz,
		       tm.purpose_kaz, tm.purpose_rus, tm.purpose_eng, tm.modules,
		       tm.created_at, tm.updated_at, tm.deleted_at
		FROM teaching_methods tm
		INNER JOIN kmj_plan_methods pm ON tm.id = pm.method_id
		WHERE pm.plan_id = $1 AND tm.deleted_at IS NULL
		ORDER BY pm.sort_order ASC
	`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []domain.TeachingMethod
	for rows.Next() {
		var method domain.TeachingMethod
		err := rows.Scan(
			&method.ID, &method.NameKaz, &method.NameRus, &method.NameEng, &method.TypeKaz,
			&method.PurposeKaz, &method.PurposeRus, &method.PurposeEng, &method.Modules,
			&method.CreatedAt, &method.UpdatedAt, &method.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, method)
	}

	if methods == nil {
		methods = []domain.TeachingMethod{}
	}

	return methods, rows.Err()
}
