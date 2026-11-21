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

// ArticleGenerationMessage is the queue payload sent from service → worker.
type ArticleGenerationMessage struct {
	ArticleID   uuid.UUID `json:"article_id"`
	UserID      uuid.UUID `json:"user_id"`
	Subject     string    `json:"subject"`
	Grade       int       `json:"grade"`
	Topic       string    `json:"topic"`
	Language    string    `json:"language"`
	Description string    `json:"description"`
	Style       string    `json:"style"`
	Complexity  string    `json:"complexity"`
}

type PaginatedArticles struct {
	Items      []domain.Article `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

type ArticleService interface {
	Create(ctx context.Context, req *domain.CreateArticleRequest, userID uuid.UUID, schoolID *int64) (*domain.Article, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Article, error)
	List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedArticles, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type articleService struct {
	repo             repository.ArticleRepository
	subscriptionRepo domain.SubscriptionRepository
	channel          *amqp.Channel
}

func NewArticleService(
	repo repository.ArticleRepository,
	subscriptionRepo domain.SubscriptionRepository,
	channel *amqp.Channel,
) ArticleService {
	return &articleService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		channel:          channel,
	}
}

func (s *articleService) Create(ctx context.Context, req *domain.CreateArticleRequest, userID uuid.UUID, schoolID *int64) (*domain.Article, error) {
	// 1. Check token balance
	balance, err := s.subscriptionRepo.CheckTokenBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token balance: %w", err)
	}
	if balance < domain.TokenCostArticleGeneration {
		return nil, fmt.Errorf("insufficient tokens: need %d, available %d", domain.TokenCostArticleGeneration, balance)
	}

	// 2. Consume tokens
	if err := s.subscriptionRepo.ConsumeTokens(ctx, userID, domain.TokenCostArticleGeneration, domain.ActionTypeArticleGeneration); err != nil {
		return nil, fmt.Errorf("insufficient tokens: %w", err)
	}

	a := &domain.Article{
		ID:          uuid.New(),
		OwnerID:     userID,
		SchoolID:    schoolID,
		Subject:     req.Subject,
		Grade:       req.Grade,
		Topic:       req.Topic,
		Language:    req.Language,
		Description: req.Description,
		Style:       req.Style,
		Complexity:  req.Complexity,
		Status:      domain.ArticleStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, fmt.Errorf("failed to create article record: %w", err)
	}

	// 3. Enqueue generation task
	msg := ArticleGenerationMessage{
		ArticleID:   a.ID,
		UserID:      userID,
		Subject:     req.Subject,
		Grade:       req.Grade,
		Topic:       req.Topic,
		Language:    req.Language,
		Description: req.Description,
		Style:       req.Style,
		Complexity:  req.Complexity,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return a, fmt.Errorf("failed to marshal queue message: %w", err)
	}

	err = s.channel.PublishWithContext(
		ctx,
		"",                   // exchange
		"article.generation", // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, a.ID, domain.ArticleStatusFailed, ptrStr("queue publish failed"))
		return a, fmt.Errorf("failed to enqueue generation: %w", err)
	}

	log.Printf("[Article] Generation queued for article %s", a.ID)
	return a, nil
}

func (s *articleService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Article, error) {
	ok, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("unauthorized: not owner")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) List(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedArticles, error) {
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

	return &PaginatedArticles{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *articleService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ok, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("unauthorized: not owner")
	}
	return s.repo.SoftDelete(ctx, id)
}
