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

type WhiteboardRepo struct {
	db *pgxpool.Pool
}

func NewWhiteboardRepository(db *pgxpool.Pool) domain.WhiteboardRepository {
	return &WhiteboardRepo{db: db}
}

func (r *WhiteboardRepo) CreateSession(ctx context.Context, session *domain.WhiteboardSession) error {
	query := `
		INSERT INTO whiteboard_sessions (id, teacher_id, class_id, title, status, snapshot, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.TeacherID,
		session.ClassID,
		session.Title,
		session.Status,
		session.Snapshot,
		session.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create whiteboard session: %w", err)
	}
	return nil
}

func (r *WhiteboardRepo) GetSession(ctx context.Context, id uuid.UUID) (*domain.WhiteboardSession, error) {
	query := `
		SELECT id, teacher_id, class_id, COALESCE(title, '') as title, status, snapshot, created_at, ended_at
		FROM whiteboard_sessions
		WHERE id = $1 AND deleted_at IS NULL
	`
	var s domain.WhiteboardSession
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID,
		&s.TeacherID,
		&s.ClassID,
		&s.Title,
		&s.Status,
		&s.Snapshot,
		&s.CreatedAt,
		&s.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get whiteboard session: %w", err)
	}
	return &s, nil
}

func (r *WhiteboardRepo) GetActiveSessionByTeacher(ctx context.Context, teacherID uuid.UUID) (*domain.WhiteboardSession, error) {
	query := `
		SELECT id, teacher_id, class_id, COALESCE(title, '') as title, status, snapshot, created_at, ended_at
		FROM whiteboard_sessions
		WHERE teacher_id = $1 AND status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s domain.WhiteboardSession
	err := r.db.QueryRow(ctx, query, teacherID).Scan(
		&s.ID,
		&s.TeacherID,
		&s.ClassID,
		&s.Title,
		&s.Status,
		&s.Snapshot,
		&s.CreatedAt,
		&s.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active whiteboard session: %w", err)
	}
	return &s, nil
}

func (r *WhiteboardRepo) GetActiveSessionByClass(ctx context.Context, classID int64) (*domain.WhiteboardSession, error) {
	query := `
		SELECT id, teacher_id, class_id, COALESCE(title, '') as title, status, snapshot, created_at, ended_at
		FROM whiteboard_sessions
		WHERE class_id = $1 AND status = 'active' AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var s domain.WhiteboardSession
	err := r.db.QueryRow(ctx, query, classID).Scan(
		&s.ID,
		&s.TeacherID,
		&s.ClassID,
		&s.Title,
		&s.Status,
		&s.Snapshot,
		&s.CreatedAt,
		&s.EndedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active whiteboard session for class: %w", err)
	}
	return &s, nil
}

func (r *WhiteboardRepo) UpdateSession(ctx context.Context, session *domain.WhiteboardSession) error {
	query := `
		UPDATE whiteboard_sessions
		SET status = $1, snapshot = $2, ended_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(ctx, query,
		session.Status,
		session.Snapshot,
		session.EndedAt,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update whiteboard session: %w", err)
	}
	return nil
}

func (r *WhiteboardRepo) CreateSummary(ctx context.Context, summary *domain.WhiteboardSummary) error {
	query := `
		INSERT INTO whiteboard_summaries (
			id, session_id, summary_text, summary_kk, summary_en, 
			explanation, explanation_kk, explanation_en, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}

	_, err := r.db.Exec(ctx, query,
		summary.ID,
		summary.SessionID,
		summary.SummaryText,
		summary.SummaryKK,
		summary.SummaryEN,
		summary.Explanation,
		summary.ExplanationKK,
		summary.ExplanationEN,
		summary.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create whiteboard summary: %w", err)
	}
	return nil
}

func (r *WhiteboardRepo) GetSummaryBySessionID(ctx context.Context, sessionID uuid.UUID) (*domain.WhiteboardSummary, error) {
	query := `
		SELECT 
			ws.id, ws.session_id, COALESCE(s.title, '') as title, 
			ws.summary_text, ws.summary_kk, ws.summary_en, 
			ws.explanation, ws.explanation_kk, ws.explanation_en, 
			ws.created_at
		FROM whiteboard_summaries ws
		JOIN whiteboard_sessions s ON ws.session_id = s.id
		WHERE ws.session_id = $1
	`
	var s domain.WhiteboardSummary
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&s.ID,
		&s.SessionID,
		&s.SessionTitle,
		&s.SummaryText,
		&s.SummaryKK,
		&s.SummaryEN,
		&s.Explanation,
		&s.ExplanationKK,
		&s.ExplanationEN,
		&s.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get whiteboard summary: %w", err)
	}
	return &s, nil
}

func (r *WhiteboardRepo) GetSummariesByClassID(ctx context.Context, classID int64) ([]domain.WhiteboardSummary, error) {
	query := `
		SELECT 
			ws.id, ws.session_id, COALESCE(s.title, '') as title, 
			ws.summary_text, ws.summary_kk, ws.summary_en, 
			ws.explanation, ws.explanation_kk, ws.explanation_en, 
			ws.created_at
		FROM whiteboard_summaries ws
		JOIN whiteboard_sessions s ON ws.session_id = s.id
		WHERE s.class_id = $1 AND s.deleted_at IS NULL
		ORDER BY ws.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to list whiteboard summaries: %w", err)
	}
	defer rows.Close()

	var summaries []domain.WhiteboardSummary
	for rows.Next() {
		var s domain.WhiteboardSummary
		err := rows.Scan(
			&s.ID,
			&s.SessionID,
			&s.SessionTitle,
			&s.SummaryText,
			&s.SummaryKK,
			&s.SummaryEN,
			&s.Explanation,
			&s.ExplanationKK,
			&s.ExplanationEN,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (r *WhiteboardRepo) ListSessionsByClass(ctx context.Context, classID int64) ([]domain.WhiteboardSession, error) {
	query := `
		SELECT id, teacher_id, class_id, COALESCE(title, '') as title, status, snapshot, created_at, ended_at
		FROM whiteboard_sessions
		WHERE class_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, classID)
	if err != nil {
		return nil, fmt.Errorf("failed to list whiteboard sessions: %w", err)
	}
	defer rows.Close()

	var sessions []domain.WhiteboardSession
	for rows.Next() {
		var s domain.WhiteboardSession
		if err := rows.Scan(&s.ID, &s.TeacherID, &s.ClassID, &s.Title, &s.Status, &s.Snapshot, &s.CreatedAt, &s.EndedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *WhiteboardRepo) ListSessionsByStudent(ctx context.Context, studentID uuid.UUID) ([]domain.WhiteboardSession, error) {
	query := `
		SELECT s.id, s.teacher_id, s.class_id, COALESCE(s.title, '') as title, s.status, s.snapshot, s.created_at, s.ended_at
		FROM whiteboard_sessions s
		JOIN class_students cs ON s.class_id = cs.class_id
		WHERE cs.student_id = $1 AND s.status = 'completed' AND s.deleted_at IS NULL
		ORDER BY s.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, studentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list whiteboard sessions for student: %w", err)
	}
	defer rows.Close()

	var sessions []domain.WhiteboardSession
	for rows.Next() {
		var s domain.WhiteboardSession
		if err := rows.Scan(&s.ID, &s.TeacherID, &s.ClassID, &s.Title, &s.Status, &s.Snapshot, &s.CreatedAt, &s.EndedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
