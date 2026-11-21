package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/repository"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ==================== DTOs ====================

// CreatePresentationRequest is the payload from the frontend to start presentation generation.
type CreatePresentationRequest struct {
	Subject     string `json:"subject" validate:"required,max=255"`
	Grade       int    `json:"grade" validate:"required,min=1,max=12"`
	Topic       string `json:"topic" validate:"required,max=500"`
	Language    string `json:"language" validate:"required,oneof=kaz rus eng"`
	SlideCount  int    `json:"slide_count" validate:"required,min=3,max=30"`
	Description string `json:"description" validate:"max=1000"`
	Style       string `json:"style" validate:"omitempty,oneof=minimal creative professional cartoon"`
}

// PresentationGenerationMessage is the queue payload sent from service → worker.
type PresentationGenerationMessage struct {
	PresentationID uuid.UUID `json:"presentation_id"`
	UserID         uuid.UUID `json:"user_id"`
	Subject        string    `json:"subject"`
	Grade          int       `json:"grade"`
	Topic          string    `json:"topic"`
	Language       string    `json:"language"`
	SlideCount     int       `json:"slide_count"`
	Description    string    `json:"description"`
	Style          string    `json:"style"`
}

// PaginatedPresentations wraps a paginated list response.
type PaginatedPresentations struct {
	Items      []domain.Presentation `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// ==================== Interface ====================

type PresentationService interface {
	Create(ctx context.Context, req *CreatePresentationRequest, userID uuid.UUID, schoolID *int64) (*domain.Presentation, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.PresentationWithSlides, error)
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedPresentations, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

// ==================== Implementation ====================

type presentationService struct {
	repo             repository.PresentationRepository
	subscriptionRepo domain.SubscriptionRepository
	channel          *amqp.Channel
}

func NewPresentationService(
	repo repository.PresentationRepository,
	subscriptionRepo domain.SubscriptionRepository,
	channel *amqp.Channel,
) PresentationService {
	return &presentationService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		channel:          channel,
	}
}

// Create creates a presentation record and enqueues an AI generation job.
func (s *presentationService) Create(ctx context.Context, req *CreatePresentationRequest, userID uuid.UUID, schoolID *int64) (*domain.Presentation, error) {
	// Default slide count
	slideCount := req.SlideCount
	if slideCount <= 0 {
		slideCount = 10
	}


	// Default style
	style := req.Style
	if style == "" {
		style = "minimal"
	}

	// Step 1: Проверить баланс токенов ПЕРЕД созданием записи
	balance, err := s.subscriptionRepo.CheckTokenBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token balance: %w", err)
	}
	if balance < domain.TokenCostSlideGeneration {
		return nil, fmt.Errorf("insufficient tokens: need %d, available %d", domain.TokenCostSlideGeneration, balance)
	}

	// Step 2: Атомарно списать токены (внутри транзакции с FOR UPDATE)
	if err := s.subscriptionRepo.ConsumeTokens(ctx, userID, domain.TokenCostSlideGeneration, domain.ActionTypeSlideGeneration); err != nil {
		// ConsumeTokens тоже проверяет баланс атомарно (защита от race condition)
		return nil, fmt.Errorf("insufficient tokens: %w", err)
	}
	log.Printf("[Presentation] %d токенов списано у пользователя %s", domain.TokenCostSlideGeneration, userID)

	p := &domain.Presentation{
		ID:          uuid.New(),
		OwnerID:     userID,
		SchoolID:    schoolID,
		Subject:     req.Subject,
		Grade:       req.Grade,
		Topic:       req.Topic,
		Language:    req.Language,
		SlideCount:  slideCount,
		Description: req.Description,
		Style:       style,
		Status:      domain.PresentationStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("failed to create presentation record: %w", err)
	}

	// Enqueue generation task
	msg := PresentationGenerationMessage{
		PresentationID: p.ID,
		UserID:         userID,
		Subject:        req.Subject,
		Grade:          req.Grade,
		Topic:          req.Topic,
		Language:       req.Language,
		SlideCount:     slideCount,
		Description:    req.Description,
		Style:          style,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return p, fmt.Errorf("failed to marshal queue message: %w", err)
	}

	err = s.channel.PublishWithContext(
		ctx,
		"",                        // exchange
		"presentation.generation", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		// Roll back status – don't block the user
		_ = s.repo.UpdateStatus(ctx, p.ID, domain.PresentationStatusFailed, ptrStr("queue publish failed"))
		return p, fmt.Errorf("failed to enqueue generation: %w", err)
	}

	log.Printf("[Presentation] Generation queued for presentation %s", p.ID)
	return p, nil
}

// GetByID returns a presentation with all its slides.
func (s *presentationService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.PresentationWithSlides, error) {
	ok, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unauthorized: not owner")
	}
	return s.repo.GetWithSlides(ctx, id)
}

// List returns a paginated list of the user's presentations.
func (s *presentationService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedPresentations, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	items, total, err := s.repo.ListByOwner(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &PaginatedPresentations{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// Delete soft-deletes a presentation.
func (s *presentationService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ok, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unauthorized: not owner")
	}
	return s.repo.SoftDelete(ctx, id)
}

// ptrStr is a small helper to take address of a string literal.
func ptrStr(s string) *string { return &s }
