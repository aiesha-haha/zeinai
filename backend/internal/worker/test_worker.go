package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/service"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TestWorker struct {
	repo         domain.TestRepository
	notification *service.NotificationService
	ai           ai.TextGenerator
	channel      *amqp.Channel
}

func NewTestWorker(
	repo domain.TestRepository,
	notification *service.NotificationService,
	ai ai.TextGenerator,
	channel *amqp.Channel,
) *TestWorker {
	return &TestWorker{
		repo:         repo,
		notification: notification,
		ai:           ai,
		channel:      channel,
	}
}

type AITestResponse struct {
	Questions []AITestQuestion `json:"questions"`
}

type AITestQuestion struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      string   `json:"answer"`
	Explanation string   `json:"explanation"`
}

func (w *TestWorker) Start(ctx context.Context) error {
	queue, err := w.channel.QueueDeclare(
		"tests.generate",
		true, false, false, false, nil,
	)
	if err != nil {
		return err
	}

	msgs, err := w.channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Printf("[Test Worker] Started, listening on queue: %s", queue.Name)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok { return nil }
			log.Printf("[Test Worker] Received message for test generation")
			w.processMessage(ctx, msg)
		}
	}
}

func (w *TestWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	var genMsg service.TestGenerationMessage
	if err := json.Unmarshal(msg.Body, &genMsg); err != nil {
		msg.Nack(false, false)
		return
	}

	// Generate questions
	questions, err := w.generateWithAI(ctx, &genMsg)
	if err != nil {
		log.Printf("[Test Worker] Generation failed: %v", err)
		errMsg := err.Error()
		_ = w.repo.UpdateStatus(ctx, genMsg.TestID, domain.TestStatusFailed, &errMsg)
		msg.Ack(false)
		return
	}

	// Save questions
	if err := w.repo.CreateQuestionsBatch(ctx, questions); err != nil {
		log.Printf("[Test Worker] Saving failed: %v", err)
		msg.Nack(false, true)
		return
	}

	// Complete task
	_ = w.repo.UpdateStatus(ctx, genMsg.TestID, domain.TestStatusCompleted, nil)
	log.Printf("[Test Worker] Successfully generated test %s for user %s", genMsg.TestID, genMsg.UserID)

	// Send notification
	if w.notification != nil {
		title := "Тест дайын!"
		message := fmt.Sprintf("«%s» тақырыбындағы тест сәтті жасалды.", genMsg.Topic)
		link := fmt.Sprintf("/teacher/materials/tests/%s", genMsg.TestID)
		if err := w.notification.CreateNotification(ctx, genMsg.UserID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
			log.Printf("[Test Worker] Failed to send notification: %v", err)
		}
	}

	msg.Ack(false)
}

func (w *TestWorker) generateWithAI(ctx context.Context, msg *service.TestGenerationMessage) ([]domain.TestQuestion, error) {
	prompt := w.buildPrompt(msg)

	var aiResp AITestResponse
	err := w.ai.GenerateJSON(ctx, prompt, &aiResp)
	if err != nil {
		return nil, err
	}

	var domainQuestions []domain.TestQuestion
	for i, q := range aiResp.Questions {
		explanation := q.Explanation
		domainQuestions = append(domainQuestions, domain.TestQuestion{
			ID:          uuid.New(),
			TestID:      msg.TestID,
			Question:    q.Question,
			Options:     q.Options,
			Answer:      q.Answer,
			Explanation: &explanation,
			Order:       i + 1,
		})
	}

	return domainQuestions, nil
}

func (w *TestWorker) buildPrompt(msg *service.TestGenerationMessage) string {
	typeDesc := "standard multiple choice"
	optionsRule := "Each question must have EXACTLY 4 options."
	switch msg.TestType {
	case "pisa":
		typeDesc = "PISA-style questions (focusing on real-world application, critical thinking, and reading comprehension). Start each question with a short real-life situation."
	case "unt":
		typeDesc = "UNT-style (National Testing in Kazakhstan format, highly rigorous academic questions)"
	case "olympiad":
		typeDesc = "olympiad-level problems: non-standard, require reasoning in several steps, noticeably harder than regular school tasks but solvable at this grade"
	case "game_quiz":
		typeDesc = "an educational quiz game for students: short, clear, engaging questions that fit on a game card (max 25 words each), gradually getting harder"
	case "game_true_false":
		typeDesc = "a TRUE/FALSE educational game: each item is a single statement about the topic that is either true or false"
		optionsRule = "Each item must have EXACTLY 2 options: the words for \"True\" and \"False\" in the requested language (kaz: \"Ақиқат\", \"Жалған\"; rus: \"Верно\", \"Неверно\"; eng: \"True\", \"False\"), always in that order. Mix true and false statements roughly 50/50."
	}

	return fmt.Sprintf(`
You are an expert teacher and curriculum designer. Create a test with %d questions.
The test type is: %s

Subject: %s
Topic: %s
Grade: %d
Language: %s

Requirements:
1. Generate exactly %d questions.
2. %s
3. One option must be the correct answer, specified exactly in the "answer" field.
4. Provide a brief explanation for why the answer is correct in the "explanation" field.
5. The content must be highly accurate and appropriate for the given grade, subject, and language.
6. Response MUST be in JSON format.

JSON Structure:
{
  "questions": [
    {
      "question": "Question text here...",
      "options": ["Option A", "Option B", "Option C", "Option D"],
      "answer": "Correct option text here",
      "explanation": "Explanation of the correct answer"
    }
  ]
}
`, msg.QuestionCount, typeDesc, msg.Subject, msg.Topic, msg.Grade, msg.Language, msg.QuestionCount, optionsRule)
}
