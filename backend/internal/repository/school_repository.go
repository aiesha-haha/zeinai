package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"zeinai-backend/internal/domain"
)

type SchoolRepo struct {
	db *pgxpool.Pool
}

func NewSchoolRepository(db *pgxpool.Pool) domain.SchoolRepository {
	return &SchoolRepo{db: db}
}

func (r *SchoolRepo) List(ctx context.Context, regionID int, limit, offset int) ([]domain.School, error) {
	query := `
		SELECT id, name_kz, name_ru, region_id, district_id, address, created_at
		FROM schools
		WHERE ($1 = 0 OR region_id = $1)
		ORDER BY id
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, regionID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list schools: %w", err)
	}
	defer rows.Close()

	var schools []domain.School
	for rows.Next() {
		var s domain.School
		if err := rows.Scan(&s.ID, &s.NameKZ, &s.NameRU, &s.RegionID, &s.DistrictID, &s.Address, &s.CreatedAt); err != nil {
			return nil, err
		}
		schools = append(schools, s)
	}
	return schools, nil
}

func (r *SchoolRepo) Search(ctx context.Context, q string, limit int) ([]domain.School, error) {
	query := `
		SELECT id, name_kz, name_ru, region_id, district_id, address, created_at
		FROM schools
		WHERE name_kz ILIKE '%' || $1 || '%' OR name_ru ILIKE '%' || $1 || '%'
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, q, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search schools: %w", err)
	}
	defer rows.Close()

	var schools []domain.School
	for rows.Next() {
		var s domain.School
		if err := rows.Scan(&s.ID, &s.NameKZ, &s.NameRU, &s.RegionID, &s.DistrictID, &s.Address, &s.CreatedAt); err != nil {
			return nil, err
		}
		schools = append(schools, s)
	}
	return schools, nil
}
