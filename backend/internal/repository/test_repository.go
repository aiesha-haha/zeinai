package repository

import (
	"context"
	"fmt"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type testRepository struct {
	db *pgxpool.Pool
}

func NewTestRepository(db *pgxpool.Pool) domain.TestRepository {
	return &testRepository{db: db}
}

func (r *testRepository) CreateTest(ctx context.Context, t *domain.Test) error {
	query := `
		INSERT INTO tests (
			id, owner_id, school_id, subject, grade, topic, 
			language, test_type, status
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		ctx, query,
		t.ID, t.OwnerID, t.SchoolID, t.Subject, t.Grade, t.Topic,
		t.Language, t.TestType, t.Status,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
}

func (r *testRepository) GetTestByID(ctx context.Context, id uuid.UUID) (*domain.Test, error) {
	var t domain.Test
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, test_type, status, error_message,
		       created_at, updated_at, deleted_at
		FROM tests
		WHERE id = $1 AND deleted_at IS NULL
	`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.OwnerID, &t.SchoolID, &t.Subject, &t.Grade, &t.Topic,
		&t.Language, &t.TestType, &t.Status, &t.ErrorMessage,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("test not found")
	}
	return &t, err
}

func (r *testRepository) GetTestWithQuestions(ctx context.Context, id uuid.UUID) (*domain.TestWithQuestions, error) {
	test, err := r.GetTestByID(ctx, id)
	if err != nil {
		return nil, err
	}
	questions, err := r.GetQuestionsByTestID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domain.TestWithQuestions{
		Test:      *test,
		Questions: questions,
	}, nil
}

func (r *testRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]domain.Test, int, error) {
	query := `
		SELECT id, owner_id, school_id, subject, grade, topic,
		       language, test_type, status, error_message,
		       created_at, updated_at
		FROM tests
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tests []domain.Test
	for rows.Next() {
		var t domain.Test
		err := rows.Scan(
			&t.ID, &t.OwnerID, &t.SchoolID, &t.Subject, &t.Grade, &t.Topic,
			&t.Language, &t.TestType, &t.Status, &t.ErrorMessage,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		tests = append(tests, t)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM tests WHERE owner_id = $1 AND deleted_at IS NULL`
	err = r.db.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if tests == nil {
		tests = []domain.Test{}
	}

	return tests, total, nil
}

func (r *testRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	query := `
		UPDATE tests
		SET status = $1, error_message = $2, updated_at = NOW()
		WHERE id = $3 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, status, errorMsg, id)
	return err
}

func (r *testRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE tests
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *testRepository) CreateQuestionsBatch(ctx context.Context, questions []domain.TestQuestion) error {
	if len(questions) == 0 {
		return nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO test_questions (
			id, test_id, question, options, answer, explanation, order_index
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	for _, q := range questions {
		_, err := tx.Exec(ctx, query, q.ID, q.TestID, q.Question, q.Options, q.Answer, q.Explanation, q.Order)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *testRepository) GetQuestionsByTestID(ctx context.Context, testID uuid.UUID) ([]domain.TestQuestion, error) {
	query := `
		SELECT id, test_id, question, options, answer, explanation, order_index, created_at
		FROM test_questions
		WHERE test_id = $1
		ORDER BY order_index ASC
	`
	rows, err := r.db.Query(ctx, query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []domain.TestQuestion
	for rows.Next() {
		var q domain.TestQuestion
		err := rows.Scan(&q.ID, &q.TestID, &q.Question, &q.Options, &q.Answer, &q.Explanation, &q.Order, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	if questions == nil {
		questions = []domain.TestQuestion{}
	}
	return questions, nil
}

func (r *testRepository) CreateAssignment(ctx context.Context, assignment *domain.TestAssignment) error {
	query := `
		INSERT INTO test_assignments (id, test_id, class_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING assigned_at
	`
	return r.db.QueryRow(
		ctx, query,
		assignment.ID, assignment.TestID, assignment.ClassID, assignment.Status,
	).Scan(&assignment.AssignedAt)
}

func (r *testRepository) CheckOwnership(ctx context.Context, testID, userID uuid.UUID) (bool, error) {
	var ownerID uuid.UUID
	query := `SELECT owner_id FROM tests WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRow(ctx, query, testID).Scan(&ownerID)
	if err == pgx.ErrNoRows {
		return false, fmt.Errorf("test not found")
	}
	if err != nil {
		return false, err
	}
	return ownerID == userID, nil
}
func (r *testRepository) ListAssignmentsForStudent(ctx context.Context, studentID uuid.UUID) ([]domain.StudentTestTask, error) {
	query := `
		SELECT ta.id as assignment_id, t.id as test_id, t.topic, t.subject, t.grade, t.test_type, ta.status, ta.assigned_at,
		       att.score, att.max_score, att.completed_at
		FROM tests t
		JOIN test_assignments ta ON t.id = ta.test_id
		JOIN class_students cs ON ta.class_id = cs.class_id
		LEFT JOIN test_attempts att ON ta.id = att.assignment_id AND cs.student_id = att.student_id
		WHERE cs.student_id = $1 AND t.deleted_at IS NULL AND ta.status = 'active'
		ORDER BY ta.assigned_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
 
	var tasks []domain.StudentTestTask
	for rows.Next() {
		var t domain.StudentTestTask
		err := rows.Scan(
			&t.AssignmentID, &t.TestID, &t.Topic, &t.Subject, &t.Grade, &t.TestType, &t.Status, &t.AssignedAt,
			&t.Score, &t.MaxScore, &t.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		if t.CompletedAt != nil {
			t.Status = "completed"
		}
		tasks = append(tasks, t)
	}
 
	if tasks == nil {
		tasks = []domain.StudentTestTask{}
	}
 
	return tasks, nil
}

func (r *testRepository) IsAssignedToClass(ctx context.Context, testID uuid.UUID, classID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM test_assignments WHERE test_id = $1 AND class_id = $2)`
	var exists bool
	err := r.db.QueryRow(ctx, query, testID, classID).Scan(&exists)
	return exists, err
}

func (r *testRepository) IsTestAssignedToStudent(ctx context.Context, testID, studentID uuid.UUID) (bool, error) {
	query := `
		SELECT 1 FROM test_assignments ta
		JOIN class_students cs ON ta.class_id = cs.class_id
		WHERE ta.test_id = $1 AND cs.student_id = $2 AND ta.status = 'active'
		LIMIT 1
	`
	var exists int
	err := r.db.QueryRow(ctx, query, testID, studentID).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *testRepository) GetAssignmentByID(ctx context.Context, id uuid.UUID) (*domain.TestAssignment, error) {
	var a domain.TestAssignment
	query := `SELECT id, test_id, class_id, assigned_at, due_date, status FROM test_assignments WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.TestID, &a.ClassID, &a.AssignedAt, &a.DueDate, &a.Status)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("assignment not found")
	}
	return &a, err
}

func (r *testRepository) GetAttempt(ctx context.Context, assignmentID, studentID uuid.UUID) (*domain.TestAttempt, error) {
	var a domain.TestAttempt
	query := `SELECT id, assignment_id, student_id, score, max_score, status, completed_at FROM test_attempts WHERE assignment_id = $1 AND student_id = $2`
	err := r.db.QueryRow(ctx, query, assignmentID, studentID).Scan(&a.ID, &a.AssignmentID, &a.StudentID, &a.Score, &a.MaxScore, &a.Status, &a.CompletedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &a, err
}

// GetAttemptAnswers returns question id -> student answer for one attempt.
func (r *testRepository) GetAttemptAnswers(ctx context.Context, attemptID uuid.UUID) (map[uuid.UUID]string, error) {
	rows, err := r.db.Query(ctx, `SELECT question_id, student_answer FROM test_answers WHERE attempt_id = $1`, attemptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]string)
	for rows.Next() {
		var qid uuid.UUID
		var ans string
		if err := rows.Scan(&qid, &ans); err != nil {
			return nil, err
		}
		result[qid] = ans
	}
	return result, rows.Err()
}

func (r *testRepository) SubmitTest(ctx context.Context, attempt *domain.TestAttempt, answers []domain.TestAnswer) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Save attempt
	queryAttempt := `
		INSERT INTO test_attempts (id, assignment_id, student_id, score, max_score, status, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = tx.Exec(ctx, queryAttempt, attempt.ID, attempt.AssignmentID, attempt.StudentID, attempt.Score, attempt.MaxScore, attempt.Status, attempt.CompletedAt)
	if err != nil {
		return err
	}

	// Save answers
	queryAnswer := `
		INSERT INTO test_answers (id, attempt_id, question_id, student_answer, is_correct)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, a := range answers {
		_, err = tx.Exec(ctx, queryAnswer, uuid.New(), attempt.ID, a.QuestionID, a.StudentAnswer, a.IsCorrect)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func (r *testRepository) ListAssignmentsByTeacher(ctx context.Context, teacherID uuid.UUID) ([]domain.TestAssignmentWithDetails, error) {
	query := `
		SELECT ta.id, ta.test_id, ta.class_id, ta.assigned_at, ta.due_date, ta.status,
		       t.topic as test_title, t.test_type, c.name as class_name,
		       (SELECT COUNT(*) FROM class_students WHERE class_id = ta.class_id) as student_count,
		       (SELECT COUNT(*) FROM test_attempts WHERE assignment_id = ta.id AND status = 'completed') as finished_count
		FROM test_assignments ta
		JOIN tests t ON ta.test_id = t.id
		JOIN classes c ON ta.class_id = c.id
		WHERE t.owner_id = $1 AND t.deleted_at IS NULL
		ORDER BY ta.assigned_at DESC
	`
	rows, err := r.db.Query(ctx, query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []domain.TestAssignmentWithDetails
	for rows.Next() {
		var a domain.TestAssignmentWithDetails
		err := rows.Scan(
			&a.ID, &a.TestID, &a.ClassID, &a.AssignedAt, &a.DueDate, &a.Status,
			&a.TestTitle, &a.TestType, &a.ClassName, &a.StudentCount, &a.FinishedCount,
		)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}

	if assignments == nil {
		assignments = []domain.TestAssignmentWithDetails{}
	}

	return assignments, nil
}

func (r *testRepository) GetAssignmentResults(ctx context.Context, assignmentID uuid.UUID) ([]domain.StudentResult, error) {
	query := `
		SELECT u.id as student_id, CONCAT(u.first_name, ' ', u.last_name) as student_name,
		       COALESCE(ta.score, 0) as score,
		       COALESCE(ta.max_score, 0) as max_score,
		       COALESCE(ta.status, 'not_started') as status,
		       ta.completed_at
		FROM users u
		JOIN class_students cs ON u.id = cs.student_id
		JOIN test_assignments tass ON cs.class_id = tass.class_id
		LEFT JOIN test_attempts ta ON tass.id = ta.assignment_id AND u.id = ta.student_id
		WHERE tass.id = $1
		ORDER BY u.first_name ASC, u.last_name ASC
	`
	rows, err := r.db.Query(ctx, query, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.StudentResult
	for rows.Next() {
		var res domain.StudentResult
		err := rows.Scan(
			&res.StudentID, &res.StudentName, &res.Score, &res.MaxScore, &res.Status, &res.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}

	if results == nil {
		results = []domain.StudentResult{}
	}

	return results, nil
}

func (r *testRepository) DeleteAssignment(ctx context.Context, assignmentID, teacherID uuid.UUID) error {
	// Verify ownership through the test
	query := `
		DELETE FROM test_assignments
		WHERE id = $1 AND test_id IN (SELECT id FROM tests WHERE owner_id = $2)
	`
	_, err := r.db.Exec(ctx, query, assignmentID, teacherID)
	return err
}
