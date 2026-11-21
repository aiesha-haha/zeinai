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

type ClassRepo struct {
	db *pgxpool.Pool
}

func NewClassRepository(db *pgxpool.Pool) domain.ClassRepository {
	return &ClassRepo{db: db}
}

func (r *ClassRepo) Create(ctx context.Context, class *domain.Class) error {
	query := `
		INSERT INTO classes (name, teacher_id, school_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	class.CreatedAt = time.Now()
	class.UpdatedAt = time.Now()

	err := r.db.QueryRow(ctx, query,
		class.Name,
		class.TeacherID,
		class.SchoolID,
		class.CreatedAt,
		class.UpdatedAt,
	).Scan(&class.ID)

	if err != nil {
		return fmt.Errorf("failed to create class: %w", err)
	}
	return nil
}

func (r *ClassRepo) GetByID(ctx context.Context, id int64) (*domain.Class, error) {
	query := `
		SELECT c.id, c.name, c.teacher_id, c.school_id, c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM class_students cs WHERE cs.class_id = c.id) as students_count
		FROM classes c
		WHERE c.id = $1 AND c.deleted_at IS NULL
	`
	var class domain.Class
	err := r.db.QueryRow(ctx, query, id).Scan(
		&class.ID,
		&class.Name,
		&class.TeacherID,
		&class.SchoolID,
		&class.CreatedAt,
		&class.UpdatedAt,
		&class.StudentCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get class: %w", err)
	}
	return &class, nil
}

func (r *ClassRepo) GetByTeacherID(ctx context.Context, teacherID uuid.UUID) ([]domain.Class, error) {
	query := `
		SELECT c.id, c.name, c.teacher_id, c.school_id, c.created_at, c.updated_at,
		(SELECT COUNT(*) FROM class_students cs WHERE cs.class_id = c.id) as students_count
		FROM classes c
		WHERE c.teacher_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to list classes: %w", err)
	}
	defer rows.Close()

	var classes []domain.Class
	for rows.Next() {
		var c domain.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.TeacherID, &c.SchoolID, &c.CreatedAt, &c.UpdatedAt, &c.StudentCount); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

func (r *ClassRepo) GetByStudentID(ctx context.Context, studentID uuid.UUID) ([]domain.Class, error) {
	query := `
		SELECT c.id, c.name, c.teacher_id, c.school_id, c.created_at, c.updated_at
		FROM classes c
		JOIN class_students cs ON c.id = cs.class_id
		WHERE cs.student_id = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get classes by student: %w", err)
	}
	defer rows.Close()

	var classes []domain.Class
	for rows.Next() {
		var c domain.Class
		if err := rows.Scan(&c.ID, &c.Name, &c.TeacherID, &c.SchoolID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, nil
}

func (r *ClassRepo) Update(ctx context.Context, class *domain.Class) error {
	query := `
		UPDATE classes
		SET name = $1, updated_at = NOW()
		WHERE id = $2 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query, class.Name, class.ID)
	if err != nil {
		return fmt.Errorf("failed to update class: %w", err)
	}
	return nil
}

func (r *ClassRepo) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Delete students normalization (remove assignments)
	_, err = tx.Exec(ctx, `DELETE FROM class_students WHERE class_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete class students: %w", err)
	}

	// 2. Delete invites
	_, err = tx.Exec(ctx, `DELETE FROM class_invites WHERE class_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete class invites: %w", err)
	}

	// 3. Soft delete class
	query := `UPDATE classes SET deleted_at = NOW() WHERE id = $1`
	_, err = tx.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete class: %w", err)
	}

	return tx.Commit(ctx)
}

func (r *ClassRepo) AddStudent(ctx context.Context, classID int64, studentID uuid.UUID) error {
	query := `
		INSERT INTO class_students (class_id, student_id, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (class_id, student_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, classID, studentID)
	if err != nil {
		return fmt.Errorf("failed to add student to class: %w", err)
	}
	return nil
}

func (r *ClassRepo) RemoveStudent(ctx context.Context, classID int64, studentID uuid.UUID) error {
	query := `DELETE FROM class_students WHERE class_id = $1 AND student_id = $2`
	_, err := r.db.Exec(ctx, query, classID, studentID)
	if err != nil {
		return fmt.Errorf("failed to remove student from class: %w", err)
	}
	return nil
}

func (r *ClassRepo) GetStudents(ctx context.Context, classID int64) ([]domain.User, error) {
	query := `
		SELECT u.id, u.phone, u.email, u.iin, u.first_name, u.last_name, u.middle_name, u.birth_year, u.gender, u.role, u.school_id, cs.created_at
		FROM users u
		JOIN class_students cs ON u.id = cs.student_id
		WHERE cs.class_id = $1 AND u.deleted_at IS NULL
		ORDER BY cs.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to get students: %w", err)
	}
	defer rows.Close()

	var students []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Phone, &u.Email, &u.IIN, &u.FirstName, &u.LastName, &u.MiddleName,
			&u.BirthYear, &u.Gender, &u.Role, &u.SchoolID, &u.CreatedAt,
		); err != nil {
			return nil, err
		}
		students = append(students, u)
	}
	return students, nil
}

func (r *ClassRepo) IsStudentInClass(ctx context.Context, classID int64, studentID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM class_students WHERE class_id = $1 AND student_id = $2 LIMIT 1`
	var exists int
	err := r.db.QueryRow(ctx, query, classID, studentID).Scan(&exists)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if student in class: %w", err)
	}
	return true, nil
}

func (r *ClassRepo) CreateInvite(ctx context.Context, invite *domain.ClassInvite) error {
	query := `
        INSERT INTO class_invites (class_id, code, usage_limit, usage_count, expires_at, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `
	invite.CreatedAt = time.Now()

	err := r.db.QueryRow(ctx, query,
		invite.ClassID,
		invite.Code,
		invite.UsageLimit,
		invite.UsageCount,
		invite.ExpiresAt,
		invite.CreatedAt,
	).Scan(&invite.ID)

	if err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}
	return nil
}

func (r *ClassRepo) GetInviteByCode(ctx context.Context, code string) (*domain.ClassInvite, error) {
	query := `
        SELECT id, class_id, code, usage_limit, usage_count, expires_at, created_at
        FROM class_invites
        WHERE code = $1
    `
	var invite domain.ClassInvite
	err := r.db.QueryRow(ctx, query, code).Scan(
		&invite.ID,
		&invite.ClassID,
		&invite.Code,
		&invite.UsageLimit,
		&invite.UsageCount,
		&invite.ExpiresAt,
		&invite.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}
	return &invite, nil
}

func (r *ClassRepo) GetActiveInvite(ctx context.Context, classID int64) (*domain.ClassInvite, error) {
	query := `
        SELECT id, class_id, code, usage_limit, usage_count, expires_at, created_at
        FROM class_invites
        WHERE class_id = $1 
          AND expires_at > NOW() 
          AND (usage_limit = 0 OR usage_count < usage_limit)
        ORDER BY created_at DESC
        LIMIT 1
    `
	var invite domain.ClassInvite
	err := r.db.QueryRow(ctx, query, classID).Scan(
		&invite.ID,
		&invite.ClassID,
		&invite.Code,
		&invite.UsageLimit,
		&invite.UsageCount,
		&invite.ExpiresAt,
		&invite.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to get active invite: %w", err)
	}
	return &invite, nil
}

func (r *ClassRepo) ExpireAllInvites(ctx context.Context, classID int64) error {
	query := `UPDATE class_invites SET expires_at = NOW() - INTERVAL '1 second' WHERE class_id = $1 AND expires_at > NOW()`
	_, err := r.db.Exec(ctx, query, classID)
	if err != nil {
		return fmt.Errorf("failed to expire invites: %w", err)
	}
	return nil
}

func (r *ClassRepo) IncrementInviteUsage(ctx context.Context, id int64) error {
	query := `UPDATE class_invites SET usage_count = usage_count + 1 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment invite usage: %w", err)
	}
	return nil
}

func (r *ClassRepo) GetClassStudentCount(ctx context.Context, classID int64) (int, error) {
	query := `SELECT COUNT(*) FROM class_students WHERE class_id = $1`
	var count int
	err := r.db.QueryRow(ctx, query, classID).Scan(&count)
	return count, err
}
