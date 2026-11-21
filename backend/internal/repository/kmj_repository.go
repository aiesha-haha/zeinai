package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KMJRepository interface {
	// Plan operations
	CreatePlan(ctx context.Context, plan *domain.KMJPlan) error
	GetPlanByID(ctx context.Context, planID uuid.UUID) (*domain.KMJPlan, error)
	GetPlanWithCells(ctx context.Context, planID uuid.UUID) (*domain.KMJPlanWithCells, error)
	ListPlansByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.KMJPlan, int, error)
	UpdatePlanStatus(ctx context.Context, planID uuid.UUID, status string) error
	SoftDeletePlan(ctx context.Context, planID uuid.UUID) error

	// Cell operations
	CreateCell(ctx context.Context, cell *domain.KMJCell) error
	CreateCellsBatch(ctx context.Context, cells []domain.KMJCell) error
	GetCellsByPlanID(ctx context.Context, planID uuid.UUID) ([]domain.KMJCell, error)
	UpdateCell(ctx context.Context, cellID uuid.UUID, cellValue string) error
	GetCellPlanID(ctx context.Context, cellID uuid.UUID) (uuid.UUID, error)
	DeleteCell(ctx context.Context, cellID uuid.UUID) error

	// Generation task operations
	CreateGenerationTask(ctx context.Context, task *domain.KMJGenerationTask) error
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*domain.KMJGenerationTask, error)
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, errorMsg *string) error
	CompleteTask(ctx context.Context, taskID uuid.UUID, aiResponseID *uuid.UUID) error

	// Ownership checks
	CheckPlanOwnership(ctx context.Context, planID, userID uuid.UUID) (bool, error)
}

type kmjRepository struct {
	db *pgxpool.Pool
}

func NewKMJRepository(db *pgxpool.Pool) KMJRepository {
	return &kmjRepository{db: db}
}

// ==================== Plan Operations ====================

func (r *kmjRepository) CreatePlan(ctx context.Context, plan *domain.KMJPlan) error {
	query := `
		INSERT INTO kmj_plans (
			id, owner_id, school_id, subject, grade, lesson_number, 
			lesson_topic, teacher_name, lesson_date, content_lang, 
			status, generation_task_id, national_value_id,
			learning_objective, duration_minutes
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		plan.ID, plan.OwnerID, plan.SchoolID, plan.Subject,
		plan.Grade, plan.LessonNumber, plan.LessonTopic,
		plan.TeacherName, plan.LessonDate, plan.ContentLang,
		plan.Status, plan.GenerationTaskID, plan.NationalValueID,
		plan.LearningObjective, plan.DurationMinutes,
	).Scan(&plan.CreatedAt, &plan.UpdatedAt)

	return err
}

func (r *kmjRepository) GetPlanByID(ctx context.Context, planID uuid.UUID) (*domain.KMJPlan, error) {
	var plan domain.KMJPlan
	query := `
		SELECT id, owner_id, school_id, subject, grade, lesson_number,
		       lesson_topic, teacher_name, lesson_date, content_lang,
		       status, generation_task_id, national_value_id, created_at, updated_at, deleted_at,
		       learning_objective, duration_minutes
		FROM kmj_plans
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, planID).Scan(
		&plan.ID, &plan.OwnerID, &plan.SchoolID, &plan.Subject,
		&plan.Grade, &plan.LessonNumber, &plan.LessonTopic,
		&plan.TeacherName, &plan.LessonDate, &plan.ContentLang,
		&plan.Status, &plan.GenerationTaskID, &plan.NationalValueID, &plan.CreatedAt,
		&plan.UpdatedAt, &plan.DeletedAt,
		&plan.LearningObjective, &plan.DurationMinutes,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("plan not found")
	}
	return &plan, err
}

func (r *kmjRepository) GetPlanWithCells(ctx context.Context, planID uuid.UUID) (*domain.KMJPlanWithCells, error) {
	plan, err := r.GetPlanByID(ctx, planID)
	if err != nil {
		return nil, err
	}

	cells, err := r.GetCellsByPlanID(ctx, planID)
	if err != nil {
		return nil, err
	}

	return &domain.KMJPlanWithCells{
		Plan:  *plan,
		Cells: cells,
	}, nil
}

func (r *kmjRepository) ListPlansByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.KMJPlan, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, lesson_number,
		       lesson_topic, teacher_name, lesson_date, content_lang,
		       status, generation_task_id, national_value_id, created_at, updated_at,
		       learning_objective, duration_minutes
		FROM kmj_plans
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var plans []domain.KMJPlan
	for rows.Next() {
		var plan domain.KMJPlan
		err := rows.Scan(
			&plan.ID, &plan.OwnerID, &plan.SchoolID, &plan.Subject,
			&plan.Grade, &plan.LessonNumber, &plan.LessonTopic,
			&plan.TeacherName, &plan.LessonDate, &plan.ContentLang,
			&plan.Status, &plan.GenerationTaskID, &plan.NationalValueID,
			&plan.CreatedAt, &plan.UpdatedAt,
			&plan.LearningObjective, &plan.DurationMinutes,
		)
		if err != nil {
			return nil, 0, err
		}
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM kmj_plans WHERE owner_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if plans == nil {
		plans = []domain.KMJPlan{}
	}

	return plans, total, nil
}

func (r *kmjRepository) UpdatePlanStatus(ctx context.Context, planID uuid.UUID, status string) error {
	query := `
		UPDATE kmj_plans
		SET status = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, status, planID)
	return err
}

func (r *kmjRepository) SoftDeletePlan(ctx context.Context, planID uuid.UUID) error {
	query := `
		UPDATE kmj_plans
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, planID)
	return err
}

// ==================== Cell Operations ====================

func (r *kmjRepository) CreateCell(ctx context.Context, cell *domain.KMJCell) error {
	query := `
		INSERT INTO kmj_cells (
			id, plan_id, section, cell_key, cell_value, 
			stage_column, is_ai_generated, is_edited
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		cell.ID, cell.PlanID, cell.Section, cell.CellKey,
		cell.CellValue, cell.StageColumn, cell.IsAIGenerated, cell.IsEdited,
	).Scan(&cell.CreatedAt, &cell.UpdatedAt)

	return err
}

func (r *kmjRepository) CreateCellsBatch(ctx context.Context, cells []domain.KMJCell) error {
	if len(cells) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO kmj_cells (
			id, plan_id, section, cell_key, cell_value, 
			stage_column, is_ai_generated, is_edited
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		)
		ON CONFLICT (plan_id, section, cell_key, (COALESCE(stage_column, '')))
		DO UPDATE SET
			cell_value = EXCLUDED.cell_value,
			is_ai_generated = EXCLUDED.is_ai_generated,
			updated_at = NOW()
	`

	for _, cell := range cells {
		_, err := tx.Exec(
			ctx, query,
			cell.ID, cell.PlanID, cell.Section, cell.CellKey,
			cell.CellValue, cell.StageColumn, cell.IsAIGenerated, cell.IsEdited,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *kmjRepository) GetCellsByPlanID(ctx context.Context, planID uuid.UUID) ([]domain.KMJCell, error) {
	query := `
		SELECT id, plan_id, section, cell_key, cell_value, stage_column,
		       is_ai_generated, is_edited, created_at, updated_at
		FROM kmj_cells
		WHERE plan_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []domain.KMJCell
	for rows.Next() {
		var cell domain.KMJCell
		err := rows.Scan(
			&cell.ID, &cell.PlanID, &cell.Section, &cell.CellKey,
			&cell.CellValue, &cell.StageColumn, &cell.IsAIGenerated,
			&cell.IsEdited, &cell.CreatedAt, &cell.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cells = append(cells, cell)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if cells == nil {
		cells = []domain.KMJCell{}
	}

	return cells, nil
}

func (r *kmjRepository) UpdateCell(ctx context.Context, cellID uuid.UUID, cellValue string) error {
	query := `
		UPDATE kmj_cells
		SET cell_value = $1, is_edited = true, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, cellValue, cellID)
	return err
}

// GetCellPlanID returns the plan a cell belongs to (used for ownership checks).
func (r *kmjRepository) GetCellPlanID(ctx context.Context, cellID uuid.UUID) (uuid.UUID, error) {
	var planID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT plan_id FROM kmj_cells WHERE id = $1`, cellID).Scan(&planID)
	if err == pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("cell not found")
	}
	return planID, err
}

func (r *kmjRepository) DeleteCell(ctx context.Context, cellID uuid.UUID) error {
	query := `DELETE FROM kmj_cells WHERE id = $1`
	_, err := r.db.Exec(ctx, query, cellID)
	return err
}

// ==================== Generation Task Operations ====================

func (r *kmjRepository) CreateGenerationTask(ctx context.Context, task *domain.KMJGenerationTask) error {
	query := `
		INSERT INTO kmj_generation_tasks (
			id, plan_id, user_id, school_id, prompt, language, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		task.ID, task.PlanID, task.UserID, task.SchoolID,
		task.Prompt, task.Language, task.Status,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	return err
}

func (r *kmjRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*domain.KMJGenerationTask, error) {
	var task domain.KMJGenerationTask
	query := `
		SELECT id, plan_id, user_id, school_id, prompt, language,
		       status, retry_count, error_message, ai_response_id,
		       created_at, updated_at, completed_at
		FROM kmj_generation_tasks
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&task.ID, &task.PlanID, &task.UserID, &task.SchoolID,
		&task.Prompt, &task.Language, &task.Status, &task.RetryCount,
		&task.ErrorMessage, &task.AIResponseID, &task.CreatedAt,
		&task.UpdatedAt, &task.CompletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	return &task, err
}

func (r *kmjRepository) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE kmj_generation_tasks
		SET status = $1, error_message = $2, updated_at = NOW(),
		    retry_count = retry_count + 1
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, errorMsg, taskID)
	return err
}

func (r *kmjRepository) CompleteTask(ctx context.Context, taskID uuid.UUID, aiResponseID *uuid.UUID) error {
	query := `
		UPDATE kmj_generation_tasks
		SET status = $1, ai_response_id = $2, completed_at = NOW(), updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, domain.TaskStatusCompleted, aiResponseID, taskID)
	return err
}

// ==================== Ownership ====================

func (r *kmjRepository) CheckPlanOwnership(ctx context.Context, planID, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	query := `SELECT owner_id FROM kmj_plans WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, planID).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("plan not found")
	}
	if err != nil {
		return false, err
	}

	return ownerID == userID, nil
}
