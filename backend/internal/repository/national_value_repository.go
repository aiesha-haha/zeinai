package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NationalValueRepository interface {
	Create(ctx context.Context, value *domain.NationalValue) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NationalValue, error)
	GetByMonth(ctx context.Context, month int) ([]domain.NationalValue, error)
	List(ctx context.Context) ([]domain.NationalValue, error)
	Update(ctx context.Context, id uuid.UUID, value *domain.NationalValue) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type nationalValueRepository struct {
	db *pgxpool.Pool
}

func NewNationalValueRepository(db *pgxpool.Pool) NationalValueRepository {
	return &nationalValueRepository{db: db}
}

func (r *nationalValueRepository) Create(ctx context.Context, value *domain.NationalValue) error {
	query := `
		INSERT INTO national_values (
			id, month_number, name_kaz, name_rus, name_eng,
			description_kaz, description_rus, description_eng, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		value.ID, value.MonthNumber, value.NameKaz, value.NameRus, value.NameEng,
		value.DescKaz, value.DescRus, value.DescEng, value.CreatedBy,
	).Scan(&value.CreatedAt, &value.UpdatedAt)

	return err
}

func (r *nationalValueRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NationalValue, error) {
	var value domain.NationalValue
	query := `
		SELECT id, month_number, name_kaz, name_rus, name_eng,
		       description_kaz, description_rus, description_eng,
		       created_by, created_at, updated_at, deleted_at
		FROM national_values
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&value.ID, &value.MonthNumber, &value.NameKaz, &value.NameRus, &value.NameEng,
		&value.DescKaz, &value.DescRus, &value.DescEng,
		&value.CreatedBy, &value.CreatedAt, &value.UpdatedAt, &value.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("national value not found")
	}
	return &value, err
}

func (r *nationalValueRepository) GetByMonth(ctx context.Context, month int) ([]domain.NationalValue, error) {
	query := `
		SELECT id, month_number, name_kaz, name_rus, name_eng,
		       description_kaz, description_rus, description_eng,
		       created_by, created_at, updated_at, deleted_at
		FROM national_values
		WHERE month_number = $1 AND deleted_at IS NULL
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []domain.NationalValue
	for rows.Next() {
		var value domain.NationalValue
		err := rows.Scan(
			&value.ID, &value.MonthNumber, &value.NameKaz, &value.NameRus, &value.NameEng,
			&value.DescKaz, &value.DescRus, &value.DescEng,
			&value.CreatedBy, &value.CreatedAt, &value.UpdatedAt, &value.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if values == nil {
		values = []domain.NationalValue{}
	}

	return values, rows.Err()
}

func (r *nationalValueRepository) List(ctx context.Context) ([]domain.NationalValue, error) {
	query := `
		SELECT id, month_number, name_kaz, name_rus, name_eng,
		       description_kaz, description_rus, description_eng,
		       created_by, created_at, updated_at, deleted_at
		FROM national_values
		WHERE deleted_at IS NULL
		ORDER BY month_number, created_at ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []domain.NationalValue
	for rows.Next() {
		var value domain.NationalValue
		err := rows.Scan(
			&value.ID, &value.MonthNumber, &value.NameKaz, &value.NameRus, &value.NameEng,
			&value.DescKaz, &value.DescRus, &value.DescEng,
			&value.CreatedBy, &value.CreatedAt, &value.UpdatedAt, &value.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	if values == nil {
		values = []domain.NationalValue{}
	}

	return values, rows.Err()
}

func (r *nationalValueRepository) Update(ctx context.Context, id uuid.UUID, value *domain.NationalValue) error {
	query := `
		UPDATE national_values
		SET month_number = $1, name_kaz = $2, name_rus = $3, name_eng = $4,
		    description_kaz = $5, description_rus = $6, description_eng = $7,
		    updated_at = NOW()
		WHERE id = $8 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(
		ctx, query,
		value.MonthNumber, value.NameKaz, value.NameRus, value.NameEng,
		value.DescKaz, value.DescRus, value.DescEng, id,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("national value not found")
	}

	return nil
}

func (r *nationalValueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE national_values
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("national value not found")
	}

	return nil
}
