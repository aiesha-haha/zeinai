package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TugOfWarWorker struct {
	repo         repository.TugOfWarRepository
	notification *service.NotificationService
	ai           ai.TextGenerator
	channel      *amqp.Channel
}

func NewTugOfWarWorker(
	repo repository.TugOfWarRepository,
	notification *service.NotificationService,
	ai ai.TextGenerator,
	channel *amqp.Channel,
) *TugOfWarWorker {
	return &TugOfWarWorker{
		repo:         repo,
		notification: notification,
		ai:           ai,
		channel:      channel,
	}
}

type AITugOfWarResponse struct {
	Group1 []AIQuestion `json:"group_1"`
	Group2 []AIQuestion `json:"group_2"`
}

type AIQuestion struct {
	Question string   `json:"question"`
	Options  []string `json:"options"` // Array of 4 strings
	Answer   string   `json:"answer"`
}

func (w *TugOfWarWorker) Start(ctx context.Context) error {
	queue, err := w.channel.QueueDeclare(
		"tug_of_war.generation",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	msgs, err := w.channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("[TugOfWar Worker] Started, listening on queue: %s", queue.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok { return nil }
			log.Printf("[TugOfWar Worker] Received message for material %s", msg.Body)
			w.processMessage(ctx, msg)
		}
	}
}

func (w *TugOfWarWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var genMsg service.TugOfWarGenerationMessage
	if err := json.Unmarshal(msg.Body, &genMsg); err != nil {
		msg.Nack(false, false)
		return
	}

	// Generate questions
	questions, err := w.generateWithAI(ctx, &genMsg)
	if err != nil {
		log.Printf("[TugOfWar Worker] Generation failed: %v", err)
		errMsg := err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.MaterialID, domain.TugOfWarStatusFailed, &errMsg)
		msg.Ack(false)
		return
	}

	// Save questions
	if err := w.repo.CreateQuestionsBatch(ctx, questions); err != nil {
		log.Printf("[TugOfWar Worker] Saving failed: %v", err)
		msg.Nack(false, true)
		return
	}

	// Complete task
	_ = w.repo.UpdateStatus(ctx, genMsg.MaterialID, domain.TugOfWarStatusCompleted, nil)
	log.Printf("[TugOfWar Worker] Successfully generated Tug of War %s for user %s", genMsg.MaterialID, genMsg.UserID)

	// Send notification
	if w.notification != nil {
		title := "Арқан тартыс дайын!"
		message := fmt.Sprintf("«%s» тақырыбындағы арқан тартыс ойыны сәтті жасалды.", genMsg.Topic)
		link := fmt.Sprintf("/teacher/materials/tug-of-war/%s", genMsg.MaterialID)
		if err := w.notification.CreateNotification(ctx, genMsg.UserID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
			log.Printf("[TugOfWar Worker] Failed to send notification: %v", err)
		}
	}

	msg.Ack(false)
}

func (w *TugOfWarWorker) generateWithAI(ctx context.Context, msg *service.TugOfWarGenerationMessage) ([]domain.TugOfWarQuestion, error) {
	prompt := w.buildPrompt(msg)

	var aiResp AITugOfWarResponse
	err := w.ai.GenerateJSON(ctx, prompt, &aiResp)
	if err != nil {
		return nil, err
	}

	var domainQuestions []domain.TugOfWarQuestion

	// Team 1
	for i, q := range aiResp.Group1 {
		domainQuestions = append(domainQuestions, domain.TugOfWarQuestion{
			ID:         uuid.New(),
			MaterialID: msg.MaterialID,
			GroupIndex: 0,
			Question:   q.Question,
			Options:    q.Options,
			Answer:     q.Answer,
			Order:      i + 1,
		})
	}

	// Team 2
	for i, q := range aiResp.Group2 {
		domainQuestions = append(domainQuestions, domain.TugOfWarQuestion{
			ID:         uuid.New(),
			MaterialID: msg.MaterialID,
			GroupIndex: 1,
			Question:   q.Question,
			Options:    q.Options,
			Answer:     q.Answer,
			Order:      i + 1,
		})
	}

	return domainQuestions, nil
}

func (w *TugOfWarWorker) buildPrompt(msg *service.TugOfWarGenerationMessage) string {
	return fmt.Sprintf(`
You are a professional teacher. Create a set of 20 multiple-choice questions for a "Tug of War" classroom game.
The questions will be divided into two groups (10 questions for each team).

Subject: %s
Topic: %s
Grade: %d
Language: %s

Requirements:
1. Generate EXACTLY 10 questions for "group_1" and EXACTLY 10 questions for "group_2".
2. Each question must have EXACTLY 4 options.
3. One option must be the correct answer, and it must be specified in the "answer" field.
4. The content must be appropriate for the grade and subject.
5. Response MUST be in JSON format.
6. The questions for both groups should be on the same topic but uniquely different.

JSON Structure:
{
  "group_1": [
    {
      "question": "Question text here",
      "options": ["Option 1", "Option 2", "Option 3", "Option 4"],
      "answer": "Correct option text here"
    }
  ],
  "group_2": [
    {
      "question": "Question text here",
      "options": ["Option 1", "Option 2", "Option 3", "Option 4"],
      "answer": "Correct option text here"
    }
  ]
}
`, msg.Subject, msg.Topic, msg.Grade, msg.Language)
}
