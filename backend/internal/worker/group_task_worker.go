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

type GroupTaskWorker struct {
	repo         repository.GroupTaskRepository
	notification *service.NotificationService
	ai            ai.TextGenerator
	channel       *amqp.Channel
}

func NewGroupTaskWorker(
	repo repository.GroupTaskRepository,
	notification *service.NotificationService,
	ai ai.TextGenerator,
	channel *amqp.Channel,
) *GroupTaskWorker {
	return &GroupTaskWorker{
		repo:         repo,
		notification: notification,
		ai:           ai,
		channel:      channel,
	}
}

type AIGroupTaskResponse struct {
	Questions []struct {
		Question string `json:"question"`
		Answer   string `json:"answer"`
	} `json:"questions"`
}

func (w *GroupTaskWorker) Start(ctx context.Context) error {
	queue, err := w.channel.QueueDeclare(
		"group_task.generation",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	msgs, err := w.channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("[GroupTask Worker] Started, listening on queue: %s", queue.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok { return nil }
			log.Printf("[GroupTask Worker] Received message for task %s", msg.Body)
			w.processMessage(ctx, msg)
		}
	}
}

func (w *GroupTaskWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var genMsg service.GroupTaskGenerationMessage
	if err := json.Unmarshal(msg.Body, &genMsg); err != nil {
		msg.Nack(false, false)
		return
	}

	// Generate questions
	questions, err := w.generateWithAI(ctx, &genMsg)
	if err != nil {
		log.Printf("[GroupTask Worker] Generation failed: %v", err)
		errMsg := err.Error()
		_ = w.repo.UpdateTaskStatus(ctx, genMsg.TaskID, domain.GroupTaskStatusFailed, &errMsg)
		msg.Ack(false) // Don't requeue if AI failed (could be persistent)
		return
	}

	// Save questions
	if err := w.repo.CreateQuestionsBatch(ctx, questions); err != nil {
		log.Printf("[GroupTask Worker] Saving failed: %v", err)
		msg.Nack(false, true)
		return
	}

	// Complete task
	_ = w.repo.UpdateTaskStatus(ctx, genMsg.TaskID, domain.GroupTaskStatusCompleted, nil)
	log.Printf("[GroupTask Worker] Successfully generated task %s for user %s", genMsg.TaskID, genMsg.UserID)

	// Send notification
	if w.notification != nil {
		title := "Топтық тапсырма дайын!"
		message := fmt.Sprintf("«%s» тақырыбындағы топтық тапсырма сәтті жасалды.", genMsg.Topic)
		link := fmt.Sprintf("/teacher/materials/group-tasks/%s", genMsg.TaskID)
		if err := w.notification.CreateNotification(ctx, genMsg.UserID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
			log.Printf("[GroupTask Worker] Failed to send notification: %v", err)
		}
	}

	msg.Ack(false)
}

func (w *GroupTaskWorker) generateWithAI(ctx context.Context, msg *service.GroupTaskGenerationMessage) ([]domain.GroupTaskQuestion, error) {
	prompt := w.buildPrompt(msg)

	var aiResp AIGroupTaskResponse
	err := w.ai.GenerateJSON(ctx, prompt, &aiResp)
	if err != nil {
		return nil, err
	}

	domainQuestions := make([]domain.GroupTaskQuestion, len(aiResp.Questions))
	for i, q := range aiResp.Questions {
		domainQuestions[i] = domain.GroupTaskQuestion{
			ID:       uuid.New(),
			TaskID:   msg.TaskID,
			Question: q.Question,
			Answer:   q.Answer,
			Order:    i + 1,
		}
	}

	return domainQuestions, nil
}

func (w *GroupTaskWorker) buildPrompt(msg *service.GroupTaskGenerationMessage) string {
	return fmt.Sprintf(`
You are a professional teacher. Create a set of %d questions and answers for a classroom group activity.
Subject: %s
Topic: %s
Language: %s

Requirements:
1. Generate EXACTLY %d questions.
2. Each question must have a clear, concise answer.
3. The content must be appropriate for the subject and topic.
4. Response MUST be in JSON format.

JSON Structure:
{
  "questions": [
    {
      "question": "Question text here",
      "answer": "Correct answer here"
    }
  ]
}
`, msg.CardCount, msg.Subject, msg.Topic, msg.Language, msg.CardCount)
}
