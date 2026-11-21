package service

import (
	"context"
	"zeinai-backend/internal/domain"
)

type AuditService struct {
	repo domain.AuditLogRepository
}

func NewAuditService(repo domain.AuditLogRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(ctx context.Context, log *domain.AuditLog) error {
	return s.repo.Create(ctx, log)
}

func (s *AuditService) GetLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	return s.repo.List(ctx, limit, offset)
}
