package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/ai"
	"zeinai-backend/internal/repository"
	"zeinai-backend/internal/service"
)

const (
	MaxRetries      = 3
	AITimeout       = 120 * time.Second // Increased timeout for Pro model
	ProcessingDelay = 500 * time.Millisecond
)

type AIKMJResponse struct {
	Header struct {
		Bolem              string `json:"bolem"`
		PedagogName        string `json:"pedagog_name"`
		Date               string `json:"date"`
		Grade              string `json:"grade"`
		Topic              string `json:"topic"`
		LearningObjectives any `json:"learning_objectives"`
		LessonObjectives   any `json:"lesson_objectives"`
		UnityCooperation   any `json:"unity_cooperation"`
		ParticipantsCount  any `json:"participants_count"`
		AbsentCount        any `json:"absent_count"`
	} `json:"header"`
	Stages struct {
		Start  AIKMJStage `json:"start"`
		Middle AIKMJStage `json:"middle"`
		End    AIKMJStage `json:"end"`
	} `json:"stages"`
	Footer struct {
		Assessment any `json:"assessment"`
		Homework   any `json:"homework"`
	} `json:"footer"`
}

type AIKMJStage struct {
	PedagogAction any `json:"pedagog_action"`
	StudentAction any `json:"student_action"`
	Assessment    any `json:"assessment"`
	Resources     any `json:"resources"`
}

type KMJWorker struct {
	kmjRepo      repository.KMJRepository
	notification *service.NotificationService
	ai           ai.TextGenerator
	channel      *amqp.Channel
}

func NewKMJWorker(
	kmjRepo repository.KMJRepository,
	notification *service.NotificationService,
	ai ai.TextGenerator,
	channel *amqp.Channel,
) *KMJWorker {
	return &KMJWorker{
		kmjRepo:      kmjRepo,
		notification: notification,
		ai:           ai,
		channel:      channel,
	}
}

// Start begins consuming from the queue
func (w *KMJWorker) Start(ctx context.Context) error {
	// Declare queue
	queue, err := w.channel.QueueDeclare(
		"kmj.generation", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS
	err = w.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Start consuming
	msgs, err := w.channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume: %w", err)
	}

	log.Printf("[KMJ Worker] Started consuming from queue: %s", queue.Name)

	// Process messages
	for {
		select {
		case <-ctx.Done():
			log.Println("[KMJ Worker] Shutting down...")
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				log.Println("[KMJ Worker] Channel closed")
				return nil
			}
			w.processMessage(ctx, msg)
		}
	}
}

func (w *KMJWorker) processMessage(ctx context.Context, msg amqp.Delivery) {
	log.Printf("[KMJ Worker] Received message: %s", msg.MessageId)

	var genMsg service.KMJGenerationMessage
	err := json.Unmarshal(msg.Body, &genMsg)
	if err != nil {
		log.Printf("[KMJ Worker] Failed to unmarshal message: %v", err)
		msg.Nack(false, false) // Don't requeue invalid messages
		return
	}

	// Get task
	task, err := w.kmjRepo.GetTaskByID(ctx, genMsg.TaskID)
	if err != nil {
		log.Printf("[KMJ Worker] Failed to get task: %v", err)
		msg.Nack(false, true) // Requeue
		return
	}

	// Check retry count
	if task.RetryCount >= MaxRetries {
		log.Printf("[KMJ Worker] Max retries reached for task %s", task.ID)
		errMsg := "Max retries exceeded"
		_ = w.kmjRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusFailed, &errMsg)
		_ = w.kmjRepo.UpdatePlanStatus(ctx, task.PlanID, domain.KMJStatusFailed)
		msg.Ack(false)
		return
	}

	// Update status to processing
	_ = w.kmjRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusProcessing, nil)

	// Small delay before processing
	time.Sleep(ProcessingDelay)

	// Generate with AI
	cells, err := w.generateWithAI(ctx, &genMsg)
	if err != nil {
		log.Printf("[KMJ Worker] AI generation failed: %v", err)
		errMsg := err.Error()
		_ = w.kmjRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusFailed, &errMsg)

		// Retry if not max retries
		if task.RetryCount < MaxRetries-1 {
			msg.Nack(false, true) // Requeue
			return
		}

		_ = w.kmjRepo.UpdatePlanStatus(ctx, task.PlanID, domain.KMJStatusFailed)
		msg.Ack(false)
		return
	}

	// Save cells
	err = w.kmjRepo.CreateCellsBatch(ctx, cells)
	if err != nil {
		log.Printf("[KMJ Worker] Failed to save cells: %v", err)
		errMsg := "Failed to save generated cells"
		_ = w.kmjRepo.UpdateTaskStatus(ctx, task.ID, domain.TaskStatusFailed, &errMsg)
		msg.Nack(false, true) // Requeue
		return
	}

	// Complete task
	err = w.kmjRepo.CompleteTask(ctx, task.ID, nil)
	if err != nil {
		log.Printf("[KMJ Worker] Failed to complete task: %v", err)
		msg.Nack(false, true)
		return
	}

	// Update plan status
	err = w.kmjRepo.UpdatePlanStatus(ctx, task.PlanID, domain.KMJStatusCompleted)
	if err != nil {
		log.Printf("[KMJ Worker] Failed to update plan status: %v", err)
	}

	log.Printf("[KMJ Worker] Successfully generated KMJ for plan %s", task.PlanID)
	
	// Send notification
	if w.notification != nil {
		plan, err := w.kmjRepo.GetPlanByID(ctx, task.PlanID)
		if err == nil {
			title := "ҚМЖ дайын!"
			message := fmt.Sprintf("«%s» тақырыбындағы ҚМЖ сәтті жасалды.", plan.LessonTopic)
			link := fmt.Sprintf("/teacher/materials/short-term-plan/%s", task.PlanID)
			log.Printf("[KMJ Worker] Sending notification to user %s for plan %s", task.UserID, task.PlanID)
			if err := w.notification.CreateNotification(ctx, task.UserID, nil, "System", title, message, domain.NotificationMaterialReady, link); err != nil {
				log.Printf("[KMJ Worker] Failed to send notification: %v", err)
			}
		} else {
			log.Printf("[KMJ Worker] Failed to fetch plan details for notification: %v", err)
		}
	}

	msg.Ack(false)
}

func (w *KMJWorker) generateWithAI(ctx context.Context, msg *service.KMJGenerationMessage) ([]domain.KMJCell, error) {
	// Create context with timeout
	aiCtx, cancel := context.WithTimeout(ctx, AITimeout)
	defer cancel()

	// Get original plan for preserving immutable fields
	plan, err := w.kmjRepo.GetPlanByID(ctx, msg.PlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plan: %w", err)
	}

	var aiResp AIKMJResponse
	err = w.ai.GenerateJSON(aiCtx, msg.Prompt, &aiResp)
	if err != nil {
		log.Printf("[KMJ Worker] AI generation failed or JSON parse error: %v", err)
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	// Convert AI response to KMJ cells, preserving original plan data
	cells := w.mapAIToCells(msg.PlanID, &aiResp, plan)

	return cells, nil
}

func (w *KMJWorker) mapAIToCells(planID uuid.UUID, aiResp *AIKMJResponse, plan *domain.KMJPlan) []domain.KMJCell {
	cells := []domain.KMJCell{}

	// Format lesson date from plan
	lessonDateStr := ""
	if plan.LessonDate != nil {
		lessonDateStr = plan.LessonDate.Format("02.01.2006")
	}

	// Teacher's own learning objective wins over the AI wording
	learningObjectives := anyToString(aiResp.Header.LearningObjectives)
	if strings.TrimSpace(plan.LearningObjective) != "" {
		learningObjectives = plan.LearningObjective
	}

	// Header cells
	headerMap := map[string]string{
		"bolem":               plan.Subject,
		"pedagog_name":        plan.TeacherName,
		"date":                lessonDateStr,
		"grade":               fmt.Sprintf("%d", plan.Grade),
		"topic":               plan.LessonTopic,
		"learning_objectives": learningObjectives,
		"lesson_objectives":   anyToString(aiResp.Header.LessonObjectives),
		"unity_cooperation":   anyToString(aiResp.Header.UnityCooperation),
		"participants_count":  anyToString(aiResp.Header.ParticipantsCount),
		"absent_count":        anyToString(aiResp.Header.AbsentCount),
	}

	for key, val := range headerMap {
		cells = append(cells, domain.KMJCell{
			ID:            uuid.New(),
			PlanID:        planID,
			Section:       domain.KMJSectionHeader,
			CellKey:       key,
			CellValue:     val,
			IsAIGenerated: true,
		})
	}

	// Helper for stages
	addStageCells := func(section string, key string, stage AIKMJStage) {
		actions := map[string]string{
			domain.KMJColumnPedagogAction: anyToString(stage.PedagogAction),
			domain.KMJColumnStudentAction: anyToString(stage.StudentAction),
			domain.KMJColumnAssessment:    anyToString(stage.Assessment),
			domain.KMJColumnResources:     anyToString(stage.Resources),
		}
		for col, val := range actions {
			colCopy := col
			cells = append(cells, domain.KMJCell{
				ID:            uuid.New(),
				PlanID:        planID,
				Section:       section,
				CellKey:       key,
				CellValue:     val,
				StageColumn:   &colCopy,
				IsAIGenerated: true,
			})
		}
	}

	addStageCells(domain.KMJSectionLessonStart, "", aiResp.Stages.Start)
	addStageCells(domain.KMJSectionLessonMiddle, "", aiResp.Stages.Middle)
	addStageCells(domain.KMJSectionLessonEnd, "", aiResp.Stages.End)

	// Footer cells
	footerMap := map[string]string{
		domain.KMJColumnAssessmentTotal: anyToString(aiResp.Footer.Assessment),
		domain.KMJColumnHomework:        anyToString(aiResp.Footer.Homework),
	}

	for col, val := range footerMap {
		colCopy := col
		cells = append(cells, domain.KMJCell{
			ID:            uuid.New(),
			PlanID:        planID,
			Section:       domain.KMJSectionFooter,
			CellKey:       "",
			CellValue:     val,
			StageColumn:   &colCopy,
			IsAIGenerated: true,
		})
	}

	return cells
}

func anyToString(val any) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case []any:
		res := ""
		for i, item := range v {
			if i > 0 {
				res += "\n"
			}
			res += fmt.Sprintf("- %v", item)
		}
		return res
	default:
		bytes, _ := json.Marshal(v)
		return string(bytes)
	}
}
