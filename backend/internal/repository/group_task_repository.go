package repository

import (
	"context"
	"fmt"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GroupTaskRepository interface {
	// Task operations
	CreateTask(ctx context.Context, task *domain.GroupTask) error
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*domain.GroupTask, error)
	GetTaskWithQuestions(ctx context.Context, taskID uuid.UUID) (*domain.GroupTaskWithQuestions, error)
	ListTasksByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.GroupTask, int, error)
	UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, errorMsg *string) error
	SoftDeleteTask(ctx context.Context, taskID uuid.UUID) error

	// Question operations
	CreateQuestionsBatch(ctx context.Context, questions []domain.GroupTaskQuestion) error
	GetQuestionsByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.GroupTaskQuestion, error)
	AddQuestion(ctx context.Context, q *domain.GroupTaskQuestion) error
	UpdateQuestion(ctx context.Context, q *domain.GroupTaskQuestion) error
	DeleteQuestion(ctx context.Context, questionID, taskID uuid.UUID) error

	// Session operations
	CreateSession(ctx context.Context, session *domain.GroupTaskSession) error
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.GroupTaskSession, error)
	GetSessionWithGroups(ctx context.Context, sessionID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error)
	UpdateSessionStatus(ctx context.Context, sessionID uuid.UUID, status string) error

	// Group/Scoring operations
	CreateGroupsBatch(ctx context.Context, groups []domain.GroupTaskGroup) error
	UpdateGroupScore(ctx context.Context, groupID uuid.UUID, score int) error
	AddStudentToGroup(ctx context.Context, groupID, studentID uuid.UUID) error

	// Ownership check
	CheckTaskOwnership(ctx context.Context, taskID, userID uuid.UUID) (bool, error)
}

type groupTaskRepository struct {
	db *pgxpool.Pool
}

func NewGroupTaskRepository(db *pgxpool.Pool) GroupTaskRepository {
	return &groupTaskRepository{db: db}
}

// ==================== Task Operations ====================

func (r *groupTaskRepository) CreateTask(ctx context.Context, task *domain.GroupTask) error {
	query := `
		INSERT INTO group_tasks (
			id, owner_id, school_id, subject, grade, topic, 
			card_count, language, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		task.ID, task.OwnerID, task.SchoolID, task.Subject,
		task.Grade, task.Topic, task.CardCount, task.Language,
		task.Status,
	).Scan(&task.CreatedAt, &task.UpdatedAt)

	return err
}

func (r *groupTaskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*domain.GroupTask, error) {
	var task domain.GroupTask
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       card_count, language, status, error_message,
		       created_at, updated_at, deleted_at
		FROM group_tasks
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRow(ctx, query, taskID).Scan(
		&task.ID, &task.OwnerID, &task.SchoolID, &task.Subject,
		&task.Grade, &task.Topic, &task.CardCount, &task.Language,
		&task.Status, &task.ErrorMessage, &task.CreatedAt,
		&task.UpdatedAt, &task.DeletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("group task not found")
	}
	return &task, err
}

func (r *groupTaskRepository) GetTaskWithQuestions(ctx context.Context, taskID uuid.UUID) (*domain.GroupTaskWithQuestions, error) {
	task, err := r.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	questions, err := r.GetQuestionsByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return &domain.GroupTaskWithQuestions{
		Task:      *task,
		Questions: questions,
	}, nil
}

func (r *groupTaskRepository) ListTasksByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.GroupTask, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       card_count, language, status, error_message,
		       created_at, updated_at
		FROM group_tasks
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []domain.GroupTask
	for rows.Next() {
		var task domain.GroupTask
		err := rows.Scan(
			&task.ID, &task.OwnerID, &task.SchoolID, &task.Subject,
			&task.Grade, &task.Topic, &task.CardCount, &task.Language,
			&task.Status, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, task)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM group_tasks WHERE owner_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if tasks == nil {
		tasks = []domain.GroupTask{}
	}

	return tasks, total, nil
}

func (r *groupTaskRepository) UpdateTaskStatus(ctx context.Context, taskID uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE group_tasks
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, status, errorMsg, taskID)
	return err
}

func (r *groupTaskRepository) SoftDeleteTask(ctx context.Context, taskID uuid.UUID) error {
	query := `
		UPDATE group_tasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, taskID)
	return err
}

// ==================== Question Operations ====================

func (r *groupTaskRepository) CreateQuestionsBatch(ctx context.Context, questions []domain.GroupTaskQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO group_task_questions (
			id, task_id, question, answer, difficulty, order_index
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, q := range questions {
		_, err := tx.Exec(ctx, query, q.ID, q.TaskID, q.Question, q.Answer, q.Difficulty, q.Order)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *groupTaskRepository) GetQuestionsByTaskID(ctx context.Context, taskID uuid.UUID) ([]domain.GroupTaskQuestion, error) {
	query := `
		SELECT id, task_id, question, answer, difficulty, order_index, created_at
		FROM group_task_questions
		WHERE task_id = $1
		ORDER BY order_index ASC
	`

	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.GroupTaskQuestion
	for rows.Next() {
		var q domain.GroupTaskQuestion
		err := rows.Scan(&q.ID, &q.TaskID, &q.Question, &q.Answer, &q.Difficulty, &q.Order, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	if questions == nil {
		questions = []domain.GroupTaskQuestion{}
	}

	return questions, nil
}

func (r *groupTaskRepository) AddQuestion(ctx context.Context, q *domain.GroupTaskQuestion) error {
	query := `
		INSERT INTO group_task_questions (id, task_id, question, answer, difficulty, order_index)
		VALUES ($1, $2, $3, $4, $5, (SELECT COALESCE(MAX(order_index), 0) + 1 FROM group_task_questions WHERE task_id = $2))
		RETURNING order_index, created_at
	`
	return r.db.QueryRow(ctx, query, q.ID, q.TaskID, q.Question, q.Answer, q.Difficulty).Scan(&q.Order, &q.CreatedAt)
}

func (r *groupTaskRepository) UpdateQuestion(ctx context.Context, q *domain.GroupTaskQuestion) error {
	query := `
		UPDATE group_task_questions
		SET question = $1, answer = $2, difficulty = $3
		WHERE id = $4 AND task_id = $5
	`
	_, err := r.db.Exec(ctx, query, q.Question, q.Answer, q.Difficulty, q.ID, q.TaskID)
	return err
}

func (r *groupTaskRepository) DeleteQuestion(ctx context.Context, questionID, taskID uuid.UUID) error {
	query := `DELETE FROM group_task_questions WHERE id = $1 AND task_id = $2`
	_, err := r.db.Exec(ctx, query, questionID, taskID)
	return err
}

// ==================== Session Operations ====================

func (r *groupTaskRepository) CreateSession(ctx context.Context, session *domain.GroupTaskSession) error {
	query := `
		INSERT INTO group_task_sessions (
			id, task_id, class_id, owner_id, status
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		ctx, query,
		session.ID, session.TaskID, session.ClassID,
		session.OwnerID, session.Status,
	).Scan(&session.CreatedAt, &session.UpdatedAt)

	return err
}

func (r *groupTaskRepository) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*domain.GroupTaskSession, error) {
	var s domain.GroupTaskSession
	query := `
		SELECT id, task_id, class_id, owner_id, status, created_at, updated_at, ended_at
		FROM group_task_sessions
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&s.ID, &s.TaskID, &s.ClassID, &s.OwnerID, &s.Status,
		&s.CreatedAt, &s.UpdatedAt, &s.EndedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	return &s, err
}

func (r *groupTaskRepository) GetSessionWithGroups(ctx context.Context, sessionID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error) {
	session, err := r.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	query := `SELECT id, session_id, name, score, created_at FROM group_task_groups WHERE session_id = $1`
	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var groups []domain.GroupTaskGroup
	for rows.Next() {
		var g domain.GroupTaskGroup
		err := rows.Scan(&g.ID, &g.SessionID, &g.Name, &g.Score, &g.CreatedAt)
		if err != nil {
			return nil, nil, err
		}
		groups = append(groups, g)
	}

	return session, groups, nil
}

func (r *groupTaskRepository) UpdateSessionStatus(ctx context.Context, sessionID uuid.UUID, status string) error {
	var endedAt *time.Time
	if status == domain.SessionStatusCompleted {
		t := time.Now()
		endedAt = &t
	}

	query := `
		UPDATE group_task_sessions
		SET status = $1, ended_at = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(ctx, query, status, endedAt, sessionID)
	return err
}

// ==================== Group/Scoring Operations ====================

func (r *groupTaskRepository) CreateGroupsBatch(ctx context.Context, groups []domain.GroupTaskGroup) error {
	if len(groups) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO group_task_groups (id, session_id, name, score) VALUES ($1, $2, $3, $4)`

	for _, g := range groups {
		_, err := tx.Exec(ctx, query, g.ID, g.SessionID, g.Name, g.Score)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *groupTaskRepository) UpdateGroupScore(ctx context.Context, groupID uuid.UUID, score int) error {
	query := `UPDATE group_task_groups SET score = score + $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, score, groupID)
	return err
}

func (r *groupTaskRepository) AddStudentToGroup(ctx context.Context, groupID, studentID uuid.UUID) error {
	query := `INSERT INTO group_task_group_students (group_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, groupID, studentID)
	return err
}

// ==================== Ownership Check ====================

func (r *groupTaskRepository) CheckTaskOwnership(ctx context.Context, taskID, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	query := `SELECT owner_id FROM group_tasks WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, taskID).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("task not found")
	}
	if err != nil {
		return false, err
	}

	return ownerID == userID, nil
}
