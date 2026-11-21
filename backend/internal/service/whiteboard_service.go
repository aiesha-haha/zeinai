package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type whiteboardService struct {
	repo                domain.WhiteboardRepository
	userRepo            domain.UserRepository
	notificationService *NotificationService
	classService        *ClassService
	channel             *amqp.Channel
}

func NewWhiteboardService(
	repo domain.WhiteboardRepository,
	userRepo domain.UserRepository,
	notificationService *NotificationService,
	classService *ClassService,
	channel *amqp.Channel,
) domain.WhiteboardService {
	return &whiteboardService{
		repo:                repo,
		userRepo:            userRepo,
		notificationService: notificationService,
		classService:        classService,
		channel:             channel,
	}
}

func (s *whiteboardService) StartSession(ctx context.Context, teacherID uuid.UUID, classID int64, title string) (*domain.WhiteboardSession, error) {
	// Check if there is already an active session for this teacher
	active, err := s.repo.GetActiveSessionByTeacher(ctx, teacherID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		// If it's the same class, return the active session. 
		// If different class, maybe end the old one or error out.
		if active.ClassID == classID {
			return active, nil
		}
		// End old one automatically or return error? Let's return error for now.
		return nil, fmt.Errorf("teacher already has an active session in class %d", active.ClassID)
	}

	if title == "" {
		title = fmt.Sprintf("Lesson - %s", time.Now().Format("02.01.2006 15:04"))
	}

	session := &domain.WhiteboardSession{
		ID:        uuid.New(),
		Title:     title,
		TeacherID: teacherID,
		ClassID:   classID,
		Status:    domain.WhiteboardStatusActive,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, err
	}

	// Notify students in the class
	u, _ := s.userRepo.GetByID(ctx, teacherID)
	senderName := "Teacher"
	if u != nil {
		senderName = u.FirstName + " " + u.LastName
	}

	go func() {
		// Use Background context for notifications to avoid cancellation when request ends
		bgCtx := context.Background()
		students, err := s.classService.ListStudents(bgCtx, classID)
		if err != nil {
			log.Printf("[Whiteboard] Failed to list students for notification: %v", err)
			return
		}

		log.Printf("[Whiteboard] Sending start notifications to %d students for class %d", len(students), classID)
		for _, student := range students {
			err := s.notificationService.CreateNotification(
				bgCtx,
				student.ID,
				&teacherID,
				senderName,
				"notif_whiteboard_started_title",
				"notif_whiteboard_started_msg",
				domain.NotificationWhiteboardStarted,
				fmt.Sprintf("/student/whiteboard/%s", session.ID),
			)
			if err != nil {
				log.Printf("[Whiteboard] Failed to send notification to student %s: %v", student.ID, err)
			}
		}
	}()

	return session, nil
}

func (s *whiteboardService) EndSession(ctx context.Context, sessionID uuid.UUID, snapshot interface{}) error {
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("session not found")
	}

	if session.Status == domain.WhiteboardStatusCompleted {
		return nil // Already ended
	}

	session.Status = domain.WhiteboardStatusCompleted
	endTime := time.Now()
	session.EndedAt = &endTime
	session.Snapshot = snapshot

	if err := s.repo.UpdateSession(ctx, session); err != nil {
		return err
	}

	// Queue summary generation
	msg := whiteboardSummaryMessage{
		SessionID: session.ID,
		TeacherID: session.TeacherID,
		ClassID:   session.ClassID,
		Title:     session.Title,
		Snapshot:  snapshot,
	}

	msgBytes, _ := json.Marshal(msg)
	err = s.channel.PublishWithContext(
		ctx,
		"",                     // exchange
		"whiteboard.summary",   // routing key
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		log.Printf("[Whiteboard] Failed to queue summary for session %s: %v", sessionID, err)
		// We don't return error because session is already marked as ended in DB
	}

	return nil
}

func (s *whiteboardService) GetSession(ctx context.Context, sessionID uuid.UUID) (*domain.WhiteboardSession, error) {
	return s.repo.GetSession(ctx, sessionID)
}

func (s *whiteboardService) GetActiveSession(ctx context.Context, userID uuid.UUID, role domain.Role) (*domain.WhiteboardSession, error) {
	if role == domain.RoleTeacher || role == domain.RoleAdmin {
		return s.repo.GetActiveSessionByTeacher(ctx, userID)
	}

	// For students, find ANY active session in their enrolled classes
	classes, err := s.classService.GetStudentClasses(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, class := range classes {
		session, err := s.repo.GetActiveSessionByClass(ctx, class.ID)
		if err != nil {
			log.Printf("[Whiteboard] Error checking active session for class %d: %v", class.ID, err)
			continue
		}
		if session != nil {
			return session, nil
		}
	}

	return nil, nil // No active session found for student
}

func (s *whiteboardService) GetHistory(ctx context.Context, classID int64) ([]domain.WhiteboardSession, error) {
	return s.repo.ListSessionsByClass(ctx, classID)
}

func (s *whiteboardService) GetStudentHistory(ctx context.Context, studentID uuid.UUID) ([]domain.WhiteboardSession, error) {
	return s.repo.ListSessionsByStudent(ctx, studentID)
}

func (s *whiteboardService) GetSummary(ctx context.Context, sessionID uuid.UUID) (*domain.WhiteboardSummary, error) {
	return s.repo.GetSummaryBySessionID(ctx, sessionID)
}

type whiteboardSummaryMessage struct {
	SessionID uuid.UUID   `json:"session_id"`
	TeacherID uuid.UUID   `json:"teacher_id"`
	ClassID   int64       `json:"class_id"`
	Title     string      `json:"title"`
	Snapshot  interface{} `json:"snapshot"`
}
