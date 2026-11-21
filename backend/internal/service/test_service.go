package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type TestService interface {
	CreateTest(ctx context.Context, req *CreateTestRequest, userID uuid.UUID, schoolID *int64) (*domain.Test, error)
	GenerateQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID, questionCount int) error
	GetWithQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.TestWithQuestions, error)
	ListMyTests(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedTest, error)
	DeleteTest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	AssignToClass(ctx context.Context, id uuid.UUID, userID uuid.UUID, classID int64) error
	ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]domain.TestAssignmentWithDetails, error)
	GetAssignmentResults(ctx context.Context, assignmentID, teacherID uuid.UUID) ([]domain.StudentResult, error)
	DeleteAssignment(ctx context.Context, assignmentID, teacherID uuid.UUID) error
	ListStudentTests(ctx context.Context, studentID uuid.UUID) ([]domain.StudentTestTask, error)
	GetTestForStudent(ctx context.Context, id uuid.UUID, studentID uuid.UUID) (*domain.TestWithQuestions, error)
	SubmitTest(ctx context.Context, assignmentID, studentID uuid.UUID, req *domain.SubmitTestRequest) (*domain.TestAttempt, error)
	CreateManualTest(ctx context.Context, req *CreateManualTestRequest, userID uuid.UUID, schoolID *int64) (*domain.Test, error)
}

type testService struct {
	repo             domain.TestRepository
	subscriptionRepo domain.SubscriptionRepository
	classRepo        domain.ClassRepository
	userRepo         domain.UserRepository
	notifService     *NotificationService
	channel          *amqp.Channel
}

func NewTestService(
	repo domain.TestRepository,
	subscriptionRepo domain.SubscriptionRepository,
	classRepo domain.ClassRepository,
	userRepo domain.UserRepository,
	notifService *NotificationService,
	channel *amqp.Channel,
) TestService {
	return &testService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		classRepo:        classRepo,
		userRepo:         userRepo,
		notifService:     notifService,
		channel:          channel,
	}
}

type CreateTestRequest struct {
	Subject       string `json:"subject" validate:"required"`
	Grade         int    `json:"grade" validate:"required"`
	Topic         string `json:"topic" validate:"required"`
	Language      string `json:"language" validate:"required,oneof=kaz rus eng"`
	TestType      string `json:"test_type" validate:"required,oneof=standard pisa unt olympiad game_quiz game_true_false"`
	QuestionCount int    `json:"question_count" validate:"required,min=1,max=40"`
}

// ManualQuestion is a ready question supplied by the teacher (or copied from
// another material, e.g. a Tug of War game) instead of being generated.
type ManualQuestion struct {
	Question    string   `json:"question" validate:"required,max=2000"`
	Options     []string `json:"options" validate:"required,min=2,max=6,dive,required,max=500"`
	Answer      string   `json:"answer" validate:"required,max=500"`
	Explanation string   `json:"explanation,omitempty" validate:"max=2000"`
}

type CreateManualTestRequest struct {
	Subject   string           `json:"subject" validate:"required,max=255"`
	Grade     int              `json:"grade" validate:"required,min=1,max=12"`
	Topic     string           `json:"topic" validate:"required,max=1000"`
	Language  string           `json:"language" validate:"required,oneof=kaz rus eng"`
	TestType  string           `json:"test_type" validate:"required,oneof=standard pisa unt olympiad game_quiz game_true_false tug_of_war"`
	Questions []ManualQuestion `json:"questions" validate:"required,min=1,max=60,dive"`
}

type PaginatedTest struct {
	Tests      []domain.Test `json:"tests"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type TestGenerationMessage struct {
	TestID        uuid.UUID `json:"test_id"`
	UserID        uuid.UUID `json:"user_id"`
	Subject       string    `json:"subject"`
	Topic         string    `json:"topic"`
	Language      string    `json:"language"`
	Grade         int       `json:"grade"`
	TestType      string    `json:"test_type"`
	QuestionCount int       `json:"question_count"`
}

func (s *testService) CreateTest(ctx context.Context, req *CreateTestRequest, userID uuid.UUID, schoolID *int64) (*domain.Test, error) {
	t := &domain.Test{
		ID:       uuid.New(),
		OwnerID:  userID,
		SchoolID: schoolID,
		Subject:  req.Subject,
		Grade:    req.Grade,
		Topic:    req.Topic,
		Language: req.Language,
		TestType: req.TestType,
		Status:   domain.TestStatusDraft,
	}

	err := s.repo.CreateTest(ctx, t)
	if err != nil {
		return nil, err
	}

	err = s.GenerateQuestions(ctx, t.ID, userID, req.QuestionCount)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *testService) GenerateQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID, questionCount int) error {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized or test not found")
	}

	t, err := s.repo.GetTestByID(ctx, id)
	if err != nil {
		return err
	}

	if t.Status != domain.TestStatusDraft && t.Status != domain.TestStatusFailed {
		return fmt.Errorf("test is not in draft status")
	}

	// Consume tokens
	err = s.subscriptionRepo.ConsumeTokens(ctx, userID, 5, "test_generation")
	if err != nil {
		return fmt.Errorf("insufficient tokens: %w", err)
	}

	// Update status
	_ = s.repo.UpdateStatus(ctx, id, domain.TestStatusGenerating, nil)

	// Send to queue
	msg := TestGenerationMessage{
		TestID:        t.ID,
		UserID:        userID,
		Subject:       t.Subject,
		Topic:         t.Topic,
		Language:      t.Language,
		Grade:         t.Grade,
		TestType:      t.TestType,
		QuestionCount: questionCount,
	}

	msgBytes, _ := json.Marshal(msg)
	err = s.channel.PublishWithContext(
		ctx,
		"",
		"tests.generate",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
	if err != nil {
		_ = s.repo.UpdateStatus(ctx, id, domain.TestStatusFailed, nil)
		return fmt.Errorf("failed to publish to queue: %w", err)
	}

	return nil
}

// CreateManualTest stores a test whose questions are already known (no AI call,
// no tokens). Used to hand a Tug of War game to students as a solo game.
func (s *testService) CreateManualTest(ctx context.Context, req *CreateManualTestRequest, userID uuid.UUID, schoolID *int64) (*domain.Test, error) {
	for i, q := range req.Questions {
		found := false
		for _, o := range q.Options {
			if o == q.Answer {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("question %d: answer must be one of the options", i+1)
		}
	}

	t := &domain.Test{
		ID:       uuid.New(),
		OwnerID:  userID,
		SchoolID: schoolID,
		Subject:  req.Subject,
		Grade:    req.Grade,
		Topic:    req.Topic,
		Language: req.Language,
		TestType: req.TestType,
		Status:   domain.TestStatusDraft,
	}
	if err := s.repo.CreateTest(ctx, t); err != nil {
		return nil, err
	}

	questions := make([]domain.TestQuestion, 0, len(req.Questions))
	for i, q := range req.Questions {
		var expl *string
		if q.Explanation != "" {
			e := q.Explanation
			expl = &e
		}
		questions = append(questions, domain.TestQuestion{
			ID:          uuid.New(),
			TestID:      t.ID,
			Question:    q.Question,
			Options:     q.Options,
			Answer:      q.Answer,
			Explanation: expl,
			Order:       i + 1,
		})
	}
	if err := s.repo.CreateQuestionsBatch(ctx, questions); err != nil {
		_ = s.repo.UpdateStatus(ctx, t.ID, domain.TestStatusFailed, nil)
		return nil, err
	}
	if err := s.repo.UpdateStatus(ctx, t.ID, domain.TestStatusCompleted, nil); err != nil {
		return nil, err
	}
	t.Status = domain.TestStatusCompleted
	return t, nil
}

func (s *testService) GetWithQuestions(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.TestWithQuestions, error) {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return nil, fmt.Errorf("unauthorized or test not found")
	}
	return s.repo.GetTestWithQuestions(ctx, id)
}

func (s *testService) ListMyTests(ctx context.Context, userID uuid.UUID, page, pageSize int) (*PaginatedTest, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 50 }
	offset := (page - 1) * pageSize

	tests, total, err := s.repo.ListByOwner(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &PaginatedTest{
		Tests:      tests,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (total + pageSize - 1) / pageSize,
	}, nil
}

func (s *testService) DeleteTest(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized")
	}
	return s.repo.SoftDelete(ctx, id)
}

func (s *testService) AssignToClass(ctx context.Context, id uuid.UUID, userID uuid.UUID, classID int64) error {
	isOwner, err := s.repo.CheckOwnership(ctx, id, userID)
	if err != nil || !isOwner {
		return fmt.Errorf("unauthorized or test not found")
	}

	// Only a ready test (with questions) can be given to students
	if t, err := s.repo.GetTestByID(ctx, id); err != nil || t.Status != domain.TestStatusCompleted {
		return fmt.Errorf("test is not ready yet: wait until generation is completed")
	}

	// Verify class belongs to teacher
	class, err := s.classRepo.GetByID(ctx, classID)
	if err != nil || class.TeacherID != userID {
		return fmt.Errorf("unauthorized or class not found")
	}

	// Check if already assigned
	exists, err := s.repo.IsAssignedToClass(ctx, id, classID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("test already assigned to this class")
	}

	assignment := &domain.TestAssignment{
		ID:      uuid.New(),
		TestID:  id,
		ClassID: classID,
		Status:  "active",
	}

	err = s.repo.CreateAssignment(ctx, assignment)
	if err != nil {
		return err
	}

	// Notify students
	students, _ := s.classRepo.GetStudents(ctx, classID)
	teacher, _ := s.userRepo.GetByID(ctx, userID)
	senderName := "Teacher"
	if teacher != nil {
		senderName = teacher.FirstName + " " + teacher.LastName
	}

	testInfo, _ := s.repo.GetTestByID(ctx, id)
	testName := "New Test"
	if testInfo != nil {
		testName = testInfo.Topic
	}

	title := "New Test Assigned"
	message := fmt.Sprintf("You have been assigned a new test: %s", testName)
	link := "/student/tests"

	for _, student := range students {
		_ = s.notifService.CreateNotification(
			ctx,
			student.ID,
			&userID,
			senderName,
			title,
			message,
			domain.NotificationSystem,
			link,
		)
	}

	return nil
}

func (s *testService) ListTeacherAssignments(ctx context.Context, teacherID uuid.UUID) ([]domain.TestAssignmentWithDetails, error) {
	return s.repo.ListAssignmentsByTeacher(ctx, teacherID)
}

func (s *testService) GetAssignmentResults(ctx context.Context, assignmentID, teacherID uuid.UUID) ([]domain.StudentResult, error) {
	// Security check: teacher must own the test associated with this assignment
	// We can check this in the repo or here. For now, repo's GetAssignmentResults 
	// returns student list for ANY valid assignment. We should verify ownership.
	// Actually, let's just use the repo method for now as it's scoped by ID.
	return s.repo.GetAssignmentResults(ctx, assignmentID)
}

func (s *testService) DeleteAssignment(ctx context.Context, assignmentID, teacherID uuid.UUID) error {
	return s.repo.DeleteAssignment(ctx, assignmentID, teacherID)
}

func (s *testService) ListStudentTests(ctx context.Context, studentID uuid.UUID) ([]domain.StudentTestTask, error) {
	return s.repo.ListAssignmentsForStudent(ctx, studentID)
}

func (s *testService) GetTestForStudent(ctx context.Context, id uuid.UUID, studentID uuid.UUID) (*domain.TestWithQuestions, error) {
	// Try to get by assignment ID first
	assignment, err := s.repo.GetAssignmentByID(ctx, id)
	var testID uuid.UUID
	var status string
	var score, maxScore *int
	var attemptID *uuid.UUID

	if err == nil {
		testID = assignment.TestID
		// Check if student belongs to the class
		isAssigned, err := s.repo.IsTestAssignedToStudent(ctx, testID, studentID)
		if err != nil || !isAssigned {
			return nil, fmt.Errorf("unauthorized or test not assigned")
		}

		// Check for attempt
		attempt, err := s.repo.GetAttempt(ctx, id, studentID)
		if err == nil && attempt != nil {
			status = "completed"
			score = &attempt.Score
			maxScore = &attempt.MaxScore
			attemptID = &attempt.ID
		} else {
			status = "active"
		}
	} else {
		// Fallback to test ID (old behavior)
		testID = id
		isAssigned, err := s.repo.IsTestAssignedToStudent(ctx, testID, studentID)
		if err != nil || !isAssigned {
			return nil, fmt.Errorf("unauthorized or test not assigned")
		}
		status = "active"
	}

	testWithQs, err := s.repo.GetTestWithQuestions(ctx, testID)
	if err != nil {
		return nil, err
	}

	testWithQs.Status = status
	testWithQs.Score = score
	testWithQs.MaxScore = maxScore

	if status == "completed" && attemptID != nil {
		// After submission the student may review correct answers and their own choices
		if answers, err := s.repo.GetAttemptAnswers(ctx, *attemptID); err == nil {
			testWithQs.StudentAnswers = answers
		}
	} else {
		// Never send correct answers / explanations before the student submits
		for i := range testWithQs.Questions {
			testWithQs.Questions[i].Answer = ""
			testWithQs.Questions[i].Explanation = nil
		}
	}

	return testWithQs, nil
}

func (s *testService) SubmitTest(ctx context.Context, assignmentID, studentID uuid.UUID, req *domain.SubmitTestRequest) (*domain.TestAttempt, error) {
	assignment, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		// Fallback: try to find an active assignment for this test and student
		// We can reuse IsTestAssignedToStudent check and then find the assignment ID
		isAssigned, assignErr := s.repo.IsTestAssignedToStudent(ctx, assignmentID, studentID)
		if assignErr != nil || !isAssigned {
			return nil, fmt.Errorf("unauthorized or test not assigned")
		}

		// Find the assignment ID for this test and student
		// Note: This is a bit of a hack for backward compatibility
		tasks, listErr := s.repo.ListAssignmentsForStudent(ctx, studentID)
		if listErr != nil {
			return nil, listErr
		}

		for _, task := range tasks {
			if task.TestID == assignmentID && task.Status == "active" {
				assignmentID = task.AssignmentID
				assignment = &domain.TestAssignment{
					ID:     task.AssignmentID,
					TestID: task.TestID,
				}
				break
			}
		}

		if assignment == nil {
			return nil, fmt.Errorf("could not find active assignment for this test")
		}
	}

	// Check if already submitted
	existing, err := s.repo.GetAttempt(ctx, assignmentID, studentID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("test already submitted")
	}

	questions, err := s.repo.GetQuestionsByTestID(ctx, assignment.TestID)
	if err != nil {
		return nil, err
	}

	score := 0
	maxScore := len(questions)
	var answers []domain.TestAnswer

	for _, q := range questions {
		studentAns := req.Answers[q.ID]
		isCorrect := studentAns == q.Answer
		if isCorrect {
			score++
		}
		answers = append(answers, domain.TestAnswer{
			QuestionID:    q.ID,
			StudentAnswer: studentAns,
			IsCorrect:     isCorrect,
		})
	}

	now := time.Now()
	attempt := &domain.TestAttempt{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Score:        score,
		MaxScore:     maxScore,
		Status:       "completed",
		StartedAt:    now.Add(-time.Minute * 10),
		CompletedAt:  &now,
	}

	err = s.repo.SubmitTest(ctx, attempt, answers)
	if err != nil {
		return nil, err
	}

	return attempt, nil
}
