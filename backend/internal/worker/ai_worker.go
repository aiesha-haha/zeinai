package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/platform/rabbitmq"

	"github.com/google/uuid"
)

type AIWorker struct {
	repo domain.AIRepository
	ai   ai.TextGenerator
	cfg  *config.Config
}

func NewAIWorker(repo domain.AIRepository, ai ai.TextGenerator, cfg *config.Config) *AIWorker {
	return &AIWorker{
		repo: repo,
		ai:   ai,
		cfg:  cfg,
	}
}

func (w *AIWorker) Start() {
	log.Println("Starting AI Worker...")
	ch := rabbitmq.GetChannel()
	if ch == nil {
		log.Println("RabbitMQ channel not ready, worker could not start")
		return
	}

	msgs, err := ch.Consume(
		"ai.text", // queue
		"",        // consumer
		false,     // auto-ack (we will ack manually)
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var payload map[string]string
			if err := json.Unmarshal(d.Body, &payload); err != nil {
				log.Printf("Error unmarshalling message: %v", err)
				d.Nack(false, false) // dead letter?
				continue
			}

			idStr, ok := payload["id"]
			if !ok {
				log.Println("Message missing id")
				d.Ack(false)
				continue
			}

			id, err := uuid.Parse(idStr)
			if err != nil {
				log.Printf("Invalid UUID: %v", err)
				d.Ack(false)
				continue
			}

			// Process
			err = w.processTask(id)
			if err != nil {
				log.Printf("Task failed: %v", err)
				// The task is marked failed in the DB. Acknowledge exactly once:
				// a Nack followed by an Ack is a protocol error that makes RabbitMQ
				// close the connection and silently stops every worker.
				d.Ack(false)
			} else {
				d.Ack(false)
			}
		}
	}()

	log.Println("Worker started. Waiting for messages.")
}

func (w *AIWorker) processTask(id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. Fetch Request
	req, err := w.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to fetch request: %w", err)
	}

	// Update status to processing
	req.Status = domain.AIStatusProcessing
	if err := w.repo.Update(ctx, req); err != nil {
		return fmt.Errorf("failed to update status to processing: %w", err)
	}

	// 2. Generation
	// Use prompt from req
	respText, err := w.ai.GenerateText(ctx, req.Prompt)
	if err != nil {
		// Mark failed
		errMsg := err.Error()
		req.Status = domain.AIStatusFailed
		req.ErrorMsg = &errMsg
		w.repo.Update(ctx, req)
		return err
	}

	// 3. Save Result
	req.Status = domain.AIStatusCompleted
	req.Response = &respText
	if err := w.repo.Update(ctx, req); err != nil {
		return fmt.Errorf("failed to save result: %w", err)
	}

	// 4. Consume tokens (estimate based on response length)
	// Rough estimate: 1 token ≈ 4 characters
	tokensConsumed := len(respText) / 4
	if tokensConsumed < 1 {
		tokensConsumed = 1 // Minimum 1 token
	}

	// Note: We need SubscriptionRepo here, but for now we'll skip this
	// TODO: Inject SubscriptionRepo into worker and consume tokens
	// subscriptionRepo.ConsumeTokens(ctx, req.UserID, tokensConsumed, "ai_text_generation")

	log.Printf("Task %s completed successfully (consumed ~%d tokens)", id, tokensConsumed)
	return nil
}
