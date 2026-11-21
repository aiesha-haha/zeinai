package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/repository"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type GroupTaskService interface {
	CreateDraft(ctx context.Context, req *CreateGroupTaskRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.GroupTask, error)
	GenerateQuestions(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error
	GetTaskWithQuestions(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.GroupTaskWithQuestions, error)
	ListMyTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedGroupTasks, error)
	DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error

	// Question editing
	AddQuestion(ctx context.Context, taskID, userID uuid.UUID, req *QuestionRequest) (*domain.GroupTaskQuestion, error)
	UpdateQuestion(ctx context.Context, taskID, questionID, userID uuid.UUID, req *QuestionRequest) error
	DeleteQuestion(ctx context.Context, taskID, questionID, userID uuid.UUID) error

	// Gameplay
	StartSession(ctx context.Context, req *StartSessionRequest, userID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error)
	AwardPoint(ctx context.Context, sessionID, groupID, userID uuid.UUID) error
	GetSession(ctx context.Context, sessionID, userID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error)
}

type groupTaskService struct {
	repo             repository.GroupTaskRepository
	classRepo        domain.ClassRepository
	subscriptionRepo domain.SubscriptionRepository
	channel          *amqp.Channel
}

func NewGroupTaskService(
	repo repository.GroupTaskRepository,
	classRepo domain.ClassRepository,
	subscriptionRepo domain.SubscriptionRepository,
	channel *amqp.Channel,
) GroupTaskService {
	return &groupTaskService{
		repo:             repo,
		classRepo:        classRepo,
		subscriptionRepo: subscriptionRepo,
		channel:          channel,
	}
}

// ==================== DTOs ====================

type CreateGroupTaskRequest struct {
	Subject   string `json:"subject" validate:"required"`
	Grade     int    `json:"grade" validate:"required"`
	Topic     string `json:"topic" validate:"required"`
	CardCount int    `json:"card_count" validate:"required,min=4,max=20"`
	Language  string `json:"language" validate:"required,oneof=kaz rus eng"`
}

type PaginatedGroupTasks struct {
	Tasks      []domain.GroupTask `json:"tasks"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

type StartSessionRequest struct {
	TaskID     uuid.UUID `json:"task_id" validate:"required"`
	ClassID    *int64    `json:"class_id,omitempty"`
	GroupCount int       `json:"group_count" validate:"required,min=2,max=10"`
}

type QuestionRequest struct {
	Question   string `json:"question" validate:"required"`
	Answer     string `json:"answer" validate:"required"`
	Difficulty string `json:"difficulty,omitempty" validate:"omitempty,oneof=easy medium hard"`
}

type GroupTaskGenerationMessage struct {
	TaskID    uuid.UUID `json:"task_id"`
	UserID    uuid.UUID `json:"user_id"`
	Subject   string    `json:"subject"`
	Topic     string    `json:"topic"`
	CardCount int       `json:"card_count"`
	Language  string    `json:"language"`
}

// ==================== Implementation ====================

func (s *groupTaskService) CreateDraft(ctx context.Context, req *CreateGroupTaskRequest, userID uuid.UUID, schoolID *uuid.UUID) (*domain.GroupTask, error) {
	task := &domain.GroupTask{
		ID:        uuid.New(),
		OwnerID:   userID,
		SchoolID:  schoolID,
		Subject:   req.Subject,
		Grade:     req.Grade,
		Topic:     req.Topic,
		CardCount: req.CardCount,
		Language:  req.Language,
		Status:    domain.GroupTaskStatusDraft,
	}

	err := s.repo.CreateTask(ctx, task)
	return task, err
}

func (s *groupTaskService) GenerateQuestions(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized or task not found")
	}

	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != domain.GroupTaskStatusDraft && task.Status != domain.GroupTaskStatusFailed {
		return fmt.Errorf("task is not in draft status")
	}

	// Consume tokens (e.g., 4 tokens for generating game questions)
	err = s.subscriptionRepo.ConsumeTokens(ctx, userID, 4, "group_task_generation")
	if err != nil {
		return fmt.Errorf("insufficient tokens: %w", err)
	}

	// Update status
	_ = s.repo.UpdateTaskStatus(ctx, taskID, domain.GroupTaskStatusGenerating, nil)

	// Send to queue
	msg := GroupTaskGenerationMessage{
		TaskID:    task.ID,
		UserID:    userID,
		Subject:   task.Subject,
		Topic:     task.Topic,
		CardCount: task.CardCount,
		Language:  task.Language,
	}

	msgBytes, _ := json.Marshal(msg)
	err = s.channel.PublishWithContext(
		ctx,
		"",                       // exchange
		"group_task.generation",  // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		_ = s.repo.UpdateTaskStatus(ctx, taskID, domain.GroupTaskStatusFailed, nil)
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	return nil
}

func (s *groupTaskService) GetTaskWithQuestions(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) (*domain.GroupTaskWithQuestions, error) {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return nil, fmt.Errorf("unauthorized or task not found")
	}
	return s.repo.GetTaskWithQuestions(ctx, taskID)
}

func (s *groupTaskService) ListMyTasks(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedGroupTasks, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 50 }
	offset := (page - 1) * pageSize

	tasks, total, err := s.repo.ListTasksByOwner(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &PaginatedGroupTasks{
		Tasks:      tasks,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}

func (s *groupTaskService) DeleteTask(ctx context.Context, taskID uuid.UUID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized")
	}
	return s.repo.SoftDeleteTask(ctx, taskID)
}

func (s *groupTaskService) AddQuestion(ctx context.Context, taskID, userID uuid.UUID, req *QuestionRequest) (*domain.GroupTaskQuestion, error) {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return nil, fmt.Errorf("unauthorized")
	}
	difficulty := req.Difficulty
	if difficulty == "" {
		difficulty = "medium"
	}
	q := &domain.GroupTaskQuestion{
		ID:         uuid.New(),
		TaskID:     taskID,
		Question:   req.Question,
		Answer:     req.Answer,
		Difficulty: difficulty,
	}
	if err := s.repo.AddQuestion(ctx, q); err != nil {
		return nil, err
	}
	return q, nil
}

func (s *groupTaskService) UpdateQuestion(ctx context.Context, taskID, questionID, userID uuid.UUID, req *QuestionRequest) error {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized")
	}
	q := &domain.GroupTaskQuestion{
		ID:         questionID,
		TaskID:     taskID,
		Question:   req.Question,
		Answer:     req.Answer,
		Difficulty: req.Difficulty,
	}
	return s.repo.UpdateQuestion(ctx, q)
}

func (s *groupTaskService) DeleteQuestion(ctx context.Context, taskID, questionID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckTaskOwnership(ctx, taskID, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized")
	}
	return s.repo.DeleteQuestion(ctx, questionID, taskID)
}

func (s *groupTaskService) StartSession(ctx context.Context, req *StartSessionRequest, userID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error) {
	// Check ownership of task
	isOwner, err := s.repo.CheckTaskOwnership(ctx, req.TaskID, userID)
	if err != nil || !isOwner {
		return nil, nil, fmt.Errorf("unauthorized or task not found")
	}

	session := &domain.GroupTaskSession{
		ID:      uuid.New(),
		TaskID:  req.TaskID,
		ClassID: req.ClassID,
		OwnerID: userID,
		Status:  domain.SessionStatusActive,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, nil, err
	}

	// Create groups
	groups := make([]domain.GroupTaskGroup, req.GroupCount)
	for i := 0; i < req.GroupCount; i++ {
		groups[i] = domain.GroupTaskGroup{
			ID:        uuid.New(),
			SessionID: session.ID,
			Name:      fmt.Sprintf("%d", i+1), // Just the number, frontend will add 'Group' or 'Топ'
			Score:     0,
		}
	}

	if err := s.repo.CreateGroupsBatch(ctx, groups); err != nil {
		return nil, nil, err
	}

	// Assign students randomly from class (only if class is provided)
	if req.ClassID != nil && *req.ClassID > 0 {
		students, err := s.classRepo.GetStudents(ctx, *req.ClassID)
		if err == nil && len(students) > 0 {
			rand.New(rand.NewSource(time.Now().UnixNano()))
			rand.Shuffle(len(students), func(i, j int) {
				students[i], students[j] = students[j], students[i]
			})

			for i, student := range students {
				groupID := groups[i%len(groups)].ID
				_ = s.repo.AddStudentToGroup(ctx, groupID, student.ID)
			}
		}
	}

	return session, groups, nil
}

func (s *groupTaskService) AwardPoint(ctx context.Context, sessionID, groupID, userID uuid.UUID) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil || session.OwnerID != userID {
		return fmt.Errorf("unauthorized or session not found")
	}

	return s.repo.UpdateGroupScore(ctx, groupID, 1)
}

func (s *groupTaskService) GetSession(ctx context.Context, sessionID, userID uuid.UUID) (*domain.GroupTaskSession, []domain.GroupTaskGroup, error) {
	session, groups, err := s.repo.GetSessionWithGroups(ctx, sessionID)
	if err != nil {
		return nil, nil, err
	}

	if session.OwnerID != userID {
		return nil, nil, fmt.Errorf("unauthorized")
	}

	return session, groups, nil
}
