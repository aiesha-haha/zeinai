package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	articleQueue      = "article.generation"
	articleTimeout    = 120 * time.Second
)

type aiArticleResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type ArticleWorker struct {
	repo         repository.ArticleRepository
	notification *service.NotificationService
	ai           ai.TextGenerator
	channel      *amqp.Channel
}

func NewArticleWorker(
	repo repository.ArticleRepository,
	notification *service.NotificationService,
	aiGen ai.TextGenerator,
	channel *amqp.Channel,
) *ArticleWorker {
	return &ArticleWorker{
		repo:         repo,
		notification: notification,
		ai:           aiGen,
		channel:      channel,
	}
}

func (w *ArticleWorker) Start(ctx context.Context) error {
	q, err := w.channel.QueueDeclare(
		articleQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("[ArticleWorker] failed to declare queue: %w", err)
	}

	if err := w.channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("[ArticleWorker] failed to set QoS: %w", err)
	}

	msgs, err := w.channel.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("[ArticleWorker] failed to consume: %w", err)
	}

	log.Printf("[ArticleWorker] Started, listening on queue: %s", q.Name)

	for {
		select {
		case <-ctx.Done():
			log.Println("[ArticleWorker] Shutting down...")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[ArticleWorker] Channel closed")
				return nil
			}
			w.processMessage(ctx, msg)
		}
	}
}

func (w *ArticleWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var genMsg service.ArticleGenerationMessage
	if err := json.Unmarshal(msg.Body, &genMsg); err != nil {
		log.Printf("[ArticleWorker] Bad message body: %v", err)
		msg.Nack(false, false)
		return
	}

	log.Printf("[ArticleWorker] Processing article %s", genMsg.ArticleID)

	// Mark as processing
	_ = w.repo.UpdateStatus(ctx, genMsg.ArticleID, domain.ArticleStatusProcessing, nil)

	// Generate with Kimi K2
	articleData, err := w.generateArticle(ctx, &genMsg)
	if err != nil {
		log.Printf("[ArticleWorker] Generation failed for %s: %v", genMsg.ArticleID, err)
		errMsg := err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.ArticleID, domain.ArticleStatusFailed, &errMsg)
		msg.Nack(false, true)
		return
	}

	// Persist result
	a, err := w.repo.GetByID(ctx, genMsg.ArticleID)
	if err != nil {
		log.Printf("[ArticleWorker] Failed to fetch article %s: %v", genMsg.ArticleID, err)
		msg.Nack(false, true)
		return
	}

	a.Title = &articleData.Title
	a.Content = &articleData.Content
	a.Status = domain.ArticleStatusCompleted

	if err := w.repo.Update(ctx, a); err != nil {
		log.Printf("[ArticleWorker] Failed to update article %s: %v", genMsg.ArticleID, err)
		errMsg := "failed to save result: " + err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.ArticleID, domain.ArticleStatusFailed, &errMsg)
		msg.Nack(false, true)
		return
	}

	log.Printf("[ArticleWorker] Article %s generated successfully", genMsg.ArticleID)

	// Send notification
	if w.notification != nil {
		title := "Мақала дайын!"
		message := fmt.Sprintf("«%s» тақырыбындағы мақалаңыз сәтті жасалды.", *a.Title)
		link := fmt.Sprintf("/teacher/materials/article/%s", genMsg.ArticleID)
		if err := w.notification.CreateNotification(ctx, a.OwnerID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
			log.Printf("[ArticleWorker] Failed to send notification: %v", err)
		}
	}

	msg.Ack(false)
}

func (w *ArticleWorker) generateArticle(ctx context.Context, msg *service.ArticleGenerationMessage) (*aiArticleResponse, error) {
	aiCtx, cancel := context.WithTimeout(ctx, articleTimeout)
	defer cancel()

	prompt := w.buildPrompt(msg)

	var aiResp aiArticleResponse
	if err := w.ai.GenerateJSON(aiCtx, prompt, &aiResp); err != nil {
		return nil, fmt.Errorf("kimi AI call failed: %w", err)
	}

	if aiResp.Title == "" || aiResp.Content == "" {
		return nil, fmt.Errorf("kimi returned empty title or content")
	}

	return &aiResp, nil
}

func (w *ArticleWorker) buildPrompt(msg *service.ArticleGenerationMessage) string {
	langMap := map[string]string{
		"kaz": "қазақ тілінде",
		"rus": "на русском языке",
		"eng": "in English",
	}
	langLabel := langMap[msg.Language]

	descBlock := ""
	if msg.Description != "" {
		descBlock = fmt.Sprintf("\nAdditional context / requirements: %s", msg.Description)
	}

	return fmt.Sprintf(`You are an expert educational content writer.

Create a high-quality, informative article for a school lesson.

Details:
- Subject: %s
- Grade: %d
- Topic: %s
- Style/Tone: %s
- Target Complexity/Length: %s
- Language: %s%s

Requirements:
1. The title should be engaging and accurate.
2. The content should be well-structured with an introduction, main body (with subheadings if appropriate), and a conclusion.
3. Use a tone suitable for grade %d students.
4. All text MUST be written %s.
5. The content should be educational and factual.

You MUST respond with ONLY a valid JSON object in this exact format (no markdown, no extra text):
{
  "title": "Article Title Here",
  "content": "Full article content with paragraphs and subheadings..."
}`,
		msg.Subject,
		msg.Grade,
		msg.Topic,
		msg.Style,
		msg.Complexity,
		langLabel,
		descBlock,
		msg.Grade,
		langLabel,
	)
}

func ptrStr(s string) *string { return &s }
