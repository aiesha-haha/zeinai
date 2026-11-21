package repository

import (
	"context"
	"fmt"
	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type STPRepository struct {
	db *pgxpool.Pool
}

func NewSTPRepository(db *pgxpool.Pool) *STPRepository {
	return &STPRepository{db: db}
}

func (r *STPRepository) GetRandomMethods(ctx context.Context, limit int) ([]domain.LessonMethod, error) {
	query := `SELECT id, name, goal, effectiveness, application, integrated_modules, created_at FROM lesson_methods ORDER BY RANDOM() LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []domain.LessonMethod
	for rows.Next() {
		var m domain.LessonMethod
		err := rows.Scan(&m.ID, &m.Name, &m.Goal, &m.Effectiveness, &m.Application, &m.IntegratedModules, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}

	return methods, nil
}

func (r *STPRepository) GetMethodsByIDs(ctx context.Context, ids []int64) ([]domain.LessonMethod, error) {
	if len(ids) == 0 {
		return []domain.LessonMethod{}, nil
	}

	query := `SELECT id, name, goal, effectiveness, application, integrated_modules, created_at FROM lesson_methods WHERE id = ANY($1)`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []domain.LessonMethod
	for rows.Next() {
		var m domain.LessonMethod
		err := rows.Scan(&m.ID, &m.Name, &m.Goal, &m.Effectiveness, &m.Application, &m.IntegratedModules, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		methods = append(methods, m)
	}

	return methods, nil
}

func (r *STPRepository) GetValueByMonth(ctx context.Context, month int) (*domain.MonthlyValue, error) {
	var val domain.MonthlyValue
	query := `SELECT id, month_number, month_name_kz, month_name_ru, value_kz, value_ru FROM monthly_values WHERE month_number = $1`
	err := r.db.QueryRow(ctx, query, month).Scan(
		&val.ID, &val.MonthNumber, &val.MonthNameKZ, &val.MonthNameRU, &val.ValueKZ, &val.ValueRU,
	)
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (r *STPRepository) CreateSTP(ctx context.Context, stp *domain.ShortTermPlan) error {
	query := `
		INSERT INTO short_term_plans (
			id, user_id, subject, grade, age_range, topic, 
			learning_objectives, lesson_objectives, selected_methods, 
			monthly_value, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
		)
	`
	_, err := r.db.Exec(ctx, query,
		stp.ID, stp.UserID, stp.Subject, stp.Grade, stp.AgeRange, stp.Topic,
		stp.LearningObjectives, stp.LessonObjectives, stp.SelectedMethods,
		stp.MonthlyValue, stp.Status, stp.CreatedAt, stp.UpdatedAt,
	)
	return err
}

func (r *STPRepository) UpdateSTP(ctx context.Context, stp *domain.ShortTermPlan) error {
	query := `
		UPDATE short_term_plans SET
			content = $1,
			additional_info = $2,
			status = $3,
			updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(ctx, query,
		stp.Content, stp.AdditionalInfo, stp.Status, stp.UpdatedAt, stp.ID,
	)
	return err
}

func (r *STPRepository) GetSTPByID(ctx context.Context, id uuid.UUID) (*domain.ShortTermPlan, error) {
	var stp domain.ShortTermPlan
	query := `SELECT id, user_id, subject, grade, age_range, topic, learning_objectives, lesson_objectives, selected_methods, monthly_value, content, additional_info, status, created_at, updated_at FROM short_term_plans WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&stp.ID, &stp.UserID, &stp.Subject, &stp.Grade, &stp.AgeRange, &stp.Topic,
		&stp.LearningObjectives, &stp.LessonObjectives, &stp.SelectedMethods,
		&stp.MonthlyValue, &stp.Content, &stp.AdditionalInfo, &stp.Status,
		&stp.CreatedAt, &stp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("stp not found")
		}
		return nil, err
	}
	return &stp, nil
}
