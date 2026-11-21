package repository

import (
	"context"
	"encoding/json"
	"zeinai-backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditRepository struct {
	db *pgxpool.Pool
}

func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	payloadJSON, err := json.Marshal(log.Payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (
			user_id, role, action, entity, entity_id, payload, source, request_id, ip_address, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`
	_, err = r.db.Exec(ctx, query,
		log.UserID,
		log.Role,
		log.Action,
		log.Entity,
		log.EntityID,
		payloadJSON,
		log.Source,
		log.RequestID,
		log.IPAddress,
	)
	return err
}

func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	query := `
		SELECT id, user_id, role, action, entity, entity_id, payload, source, request_id, ip_address, created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var payloadJSON []byte
		err := rows.Scan(
			&l.ID,
			&l.UserID,
			&l.Role,
			&l.Action,
			&l.Entity,
			&l.EntityID,
			&payloadJSON,
			&l.Source,
			&l.RequestID,
			&l.IPAddress,
			&l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(payloadJSON, &l.Payload)
		logs = append(logs, l)
	}
	return logs, nil
}
