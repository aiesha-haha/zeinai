package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/repository"
	"zeinai-backend/platform/rabbitmq"

	"github.com/google/uuid"
)

type STPService struct {
	repo *repository.STPRepository
}

func NewSTPService(repo *repository.STPRepository) *STPService {
	return &STPService{
		repo: repo,
	}
}

func (s *STPService) GetMethodsForUser(ctx context.Context) ([]domain.LessonMethod, error) {
	return s.repo.GetRandomMethods(ctx, 5)
}

func (s *STPService) GenerateSTP(ctx context.Context, userID uuid.UUID, dto *domain.STPRequestDTO) (uuid.UUID, error) {
	// 1. Get Monthly Value
	val, err := s.repo.GetValueByMonth(ctx, dto.Month)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get monthly value: %w", err)
	}

	// 2. Get Methods
	methods, err := s.repo.GetMethodsByIDs(ctx, dto.MethodIDs)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get methods: %w", err)
	}

	selectedMethodNames := []string{}
	for _, m := range methods {
		selectedMethodNames = append(selectedMethodNames, m.Name)
	}
	selectedMethodNames = append(selectedMethodNames, dto.CustomMethods...)

	// 3. Determine Age Range
	ageRange := fmt.Sprintf("%d-%d жас", dto.Grade+5, dto.Grade+6)

	// 4. Create record
	selectedMethodsJSON, _ := json.Marshal(selectedMethodNames)

	stp := &domain.ShortTermPlan{
		ID:                 uuid.New(),
		UserID:             userID,
		Subject:            dto.Subject,
		Grade:              dto.Grade,
		AgeRange:           ageRange,
		Topic:              dto.Topic,
		LearningObjectives: dto.LearningObjectives,
		LessonObjectives:   dto.LessonObjectives,
		SelectedMethods:    selectedMethodsJSON,
		MonthlyValue:       val.ValueKZ,
		Status:             domain.AIStatusPending,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	err = s.repo.CreateSTP(ctx, stp)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create stp record: %w", err)
	}

	// 5. Send to RabbitMQ
	payload := map[string]string{"id": stp.ID.String()}
	body, err := json.Marshal(payload)
	if err != nil {
		return stp.ID, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := rabbitmq.Publish("ai.stp", body); err != nil {
		log.Printf("Error publishing to RabbitMQ: %v", err)
		return stp.ID, fmt.Errorf("failed to enqueue task: %w", err)
	}

	return stp.ID, nil
}

func (s *STPService) GetSTPStatus(ctx context.Context, id uuid.UUID) (*domain.ShortTermPlan, error) {
	return s.repo.GetSTPByID(ctx, id)
}
