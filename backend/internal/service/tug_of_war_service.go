package service

import (
	"context"
	"encoding/json"
	"fmt"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/repository"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TugOfWarService interface {
	CreateDraft(ctx context.Context, req *CreateTugOfWarRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.TugOfWarMaterial, error)
	GenerateQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetWithQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.TugOfWarWithQuestions, error)
	ListMyMaterials(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedTugOfWar, error)
	DeleteMaterial(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type tugOfWarService struct {
	repo             repository.TugOfWarRepository
	subscriptionRepo domain.SubscriptionRepository
	channel          *amqp.Channel
}

// Re-checking domain interface names
func NewTugOfWarService(
	repo repository.TugOfWarRepository,
	subscriptionRepo domain.SubscriptionRepository,
	channel *amqp.Channel,
) TugOfWarService {
	return &tugOfWarService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		channel:          channel,
	}
}

// ==================== DTOs = immigrants ====================

type CreateTugOfWarRequest struct {
	Subject    string `json:"subject" validate:"required"`
	Grade      int    `json:"grade" validate:"required"`
	Topic      string `json:"topic" validate:"required"`
	Language   string `json:"language" validate:"required,oneof=kaz rus eng"`
	Group1Name string `json:"group_1_name" validate:"required"`
	Group2Name string `json:"group_2_name" validate:"required"`
}

type PaginatedTugOfWar struct {
	Materials  []domain.TugOfWarMaterial `json:"materials"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
}

type TugOfWarGenerationMessage struct {
	MaterialID uuid.UUID `json:"material_id"`
	UserID     uuid.UUID `json:"user_id"`
	Subject    string    `json:"subject"`
	Topic      string    `json:"topic"`
	Language   string    `json:"language"`
	Grade      int       `json:"grade"`
}

// ==================== Implementation ====================

func (s *tugOfWarService) CreateDraft(ctx context.Context, req *CreateTugOfWarRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.TugOfWarMaterial, error) {
	m := &domain.TugOfWarMaterial{
		ID:         uuid.New(),
		OwnerID:    userID,
		SchoolID:   schoolID,
		Subject:    req.Subject,
		Grade:      req.Grade,
		Topic:      req.Topic,
		Language:   req.Language,
		Group1Name: req.Group1Name,
		Group2Name: req.Group2Name,
		Status:     domain.TugOfWarStatusDraft,
	}

	err := s.repo.CreateMaterial(ctx, m)
	return m, err
}

func (s *tugOfWarService) GenerateQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized or material not found")
	}

	m, err := s.repo.GetMaterialByID(ctx, id)
	if err != nil {
		return err
	}

	if m.Status != domain.TugOfWarStatusDraft && m.Status != domain.TugOfWarStatusFailed {
		return fmt.Errorf("material is not in draft status")
	}

	// Consume tokens (5 tokens for generating game questions)
	err = s.subscriptionRepo.ConsumeTokens(ctx, userID, 5, "tug_of_war_generation")
	if err != nil {
		return fmt.Errorf("insufficient tokens: %w", err)
	}

	// Update status
	_ = s.repo.UpdateStatus(ctx, id, domain.TugOfWarStatusGenerating, nil)

	// Send to queue
	msg := TugOfWarGenerationMessage{
		MaterialID: m.ID,
		UserID:     userID,
		Subject:    m.Subject,
		Topic:      m.Topic,
		Language:   m.Language,
		Grade:      m.Grade,
	}

	msgBytes, _ := json.Marshal(msg)
	err = s.channel.PublishWithContext(
		ctx,
		"",                        // exchange
		"tug_of_war.generation",   // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, id, domain.TugOfWarStatusFailed, nil)
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	return nil
}

func (s *tugOfWarService) GetWithQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.TugOfWarWithQuestions, error) {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return nil, fmt.Errorf("unauthorized or material not found")
	}
	return s.repo.GetMaterialWithQuestions(ctx, id)
}

func (s *tugOfWarService) ListMyMaterials(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedTugOfWar, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 50 }
	offset := (page - 1) * pageSize

	materials, total, err := s.repo.ListByOwner(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &PaginatedTugOfWar{
		Materials:  materials,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}

func (s *tugOfWarService) DeleteMaterial(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized")
	}
	return s.repo.SoftDelete(ctx, id)
}
