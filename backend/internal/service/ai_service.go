package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"zeinai-backend/internal/domain"
	"zeinai-backend/platform/rabbitmq"
)

type AIService struct {
	repo domain.AIRepository
}

func NewAIService(repo domain.AIRepository) *AIService {
	return &AIService{repo: repo}
}

// EnqueueTextGeneration creates a request and sends it to the queue
func (s *AIService) EnqueueTextGeneration(ctx context.Context, userID uuid.UUID, prompt string) (*domain.AIRequest, error) {
	// 1. Create DB Record
	req := &domain.AIRequest{
		UserID: userID,
		Type:   domain.AIRequestTypeText,
		Prompt: prompt,
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 2. Publish to RabbitMQ
	// Payload only needs ID, but for some consumers maybe type is useful. 
	// Rule: "In queue only ID of task"
	payload := map[string]string{"id": req.ID.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := rabbitmq.Publish("ai.text", body); err != nil {
		log.Printf("Error publishing to RabbitMQ: %v", err)
		// Mark as failed in DB? Or retry? For now, just log.
		// ideally we should probably fail the request or have a transactional outbox.
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return req, nil
}

// SuggestArticleInstructions generates a prompt for instructions and enqueues it
func (s *AIService) SuggestArticleInstructions(ctx context.Context, userID uuid.UUID, subject string, grade int, topic string, language string) (*domain.AIRequest, error) {
	prompt := fmt.Sprintf(
		"Generate short, professional 'additional instructions' (in %s) for an AI that will write an educational article for %d-grade students on the topic: '%s' (Subject: %s). Return only 3-5 concise bullet points of what should be emphasized.",
		language, grade, topic, subject,
	)

	return s.EnqueueTextGeneration(ctx, userID, prompt)
}

// SuggestPresentationInstructions generates a prompt for presentation instructions and enqueues it
func (s *AIService) SuggestPresentationInstructions(ctx context.Context, userID uuid.UUID, subject string, grade int, topic string, language string) (*domain.AIRequest, error) {
	prompt := fmt.Sprintf(
		"Generate short, professional 'additional instructions' (in %s) for an AI that will create an educational presentation (slides) for %d-grade students on the topic: '%s' (Subject: %s). Return only 3-5 concise bullet points of what kind of content, structure, and visual hints should be included in the presentation.",
		language, grade, topic, subject,
	)

	return s.EnqueueTextGeneration(ctx, userID, prompt)
}

func (s *AIService) GetRequest(ctx context.Context, id uuid.UUID) (*domain.AIRequest, error) {
	return s.repo.GetByID(ctx, id)
}
