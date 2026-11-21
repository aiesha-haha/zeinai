package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type ClassService struct {
	repo         domain.ClassRepository
	userRepo     domain.UserRepository
	auditService *AuditService
	notifService *NotificationService
	whiteboardRepo domain.WhiteboardRepository
}

func NewClassService(repo domain.ClassRepository, userRepo domain.UserRepository, audit *AuditService, notif *NotificationService, whiteboardRepo domain.WhiteboardRepository) *ClassService {
	return &ClassService{
		repo:           repo,
		userRepo:       userRepo,
		auditService:  audit,
		notifService:  notif,
		whiteboardRepo: whiteboardRepo,
	}
}

func (s *ClassService) NotifyStudent(ctx context.Context, teacherID uuid.UUID, req domain.TeacherNotifyStudentRequest) error {
	// 1. Verify this student belongs to at least one class of this teacher
	// (Simple check: we list all students for this teacher or check class membership)
	// For simplicity, let's just send it if student is a student
	student, err := s.userRepo.GetByID(ctx, req.StudentID)
	if err != nil || student == nil || student.Role != domain.RoleStudent {
		return errors.New("student not found or invalid role")
	}

	teacher, _ := s.userRepo.GetByID(ctx, teacherID)
	senderName := "Teacher"
	if teacher != nil {
		senderName = teacher.FirstName + " " + teacher.LastName
	}

	return s.notifService.CreateNotification(ctx, req.StudentID, &teacherID, senderName, req.Title, req.Message, domain.NotificationSystem, "")
}

func (s *ClassService) NotifyClass(ctx context.Context, teacherID uuid.UUID, req domain.TeacherNotifyClassRequest) error {
	// 1. Verify ownership
	class, err := s.GetClass(ctx, req.ClassID, teacherID)
	if err != nil {
		return err
	}

	// 2. Get all students in class
	students, err := s.repo.GetStudents(ctx, class.ID)
	if err != nil {
		return err
	}

	// 3. Get Teacher name
	teacher, _ := s.userRepo.GetByID(ctx, teacherID)
	senderName := "Teacher"
	if teacher != nil {
		senderName = teacher.FirstName + " " + teacher.LastName
	}

	// 4. Notify all
	for _, student := range students {
		_ = s.notifService.CreateNotification(ctx, student.ID, &teacherID, senderName, req.Title, req.Message, domain.NotificationSystem, "")
	}

	return nil
}

func (s *ClassService) NotifyAllStudents(ctx context.Context, teacherID uuid.UUID, req domain.TeacherNotifyAllStudentsRequest) error {
	// 1. Get all classes of teacher
	classes, err := s.repo.GetByTeacherID(ctx, teacherID)
	if err != nil {
		return err
	}

	// 2. Get Teacher name
	teacher, _ := s.userRepo.GetByID(ctx, teacherID)
	senderName := "Teacher"
	if teacher != nil {
		senderName = teacher.FirstName + " " + teacher.LastName
	}

	// 3. Use a map to avoid duplicate notifications (student might be in multiple classes)
	sentTo := make(map[uuid.UUID]bool)

	for _, class := range classes {
		students, _ := s.repo.GetStudents(ctx, class.ID)
		for _, student := range students {
			if !sentTo[student.ID] {
				_ = s.notifService.CreateNotification(ctx, student.ID, &teacherID, senderName, req.Title, req.Message, domain.NotificationSystem, "")
				sentTo[student.ID] = true
			}
		}
	}

	return nil
}

func (s *ClassService) CreateClass(ctx context.Context, req domain.CreateClassRequest, teacherID uuid.UUID) (*domain.Class, error) {
	// Fetch teacher to get SchoolID
	teacher, err := s.userRepo.GetByID(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("failed to get teacher info: %w", err)
	}

	class := &domain.Class{
		Name:      req.Name,
		TeacherID: teacherID,
		SchoolID:  teacher.SchoolID,
	}

	if err := s.repo.Create(ctx, class); err != nil {
		return nil, err
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_CREATE",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
	})

	return class, nil
}

func (s *ClassService) GetClasses(ctx context.Context, teacherID uuid.UUID) ([]domain.Class, error) {
	return s.repo.GetByTeacherID(ctx, teacherID)
}

func (s *ClassService) GetStudentClasses(ctx context.Context, studentID uuid.UUID) ([]domain.Class, error) {
	return s.repo.GetByStudentID(ctx, studentID)
}

func (s *ClassService) GetClass(ctx context.Context, id int64, teacherID uuid.UUID) (*domain.Class, error) {
	class, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if class == nil {
		return nil, errors.New("class not found")
	}
	if class.TeacherID != teacherID {
		return nil, errors.New("unauthorized")
	}
	return class, nil
}

func (s *ClassService) UpdateClass(ctx context.Context, id int64, req domain.CreateClassRequest, teacherID uuid.UUID) (*domain.Class, error) {
	class, err := s.GetClass(ctx, id, teacherID)
	if err != nil {
		return nil, err
	}

	class.Name = req.Name
	if err := s.repo.Update(ctx, class); err != nil {
		return nil, err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_UPDATE",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
	})

	return class, nil
}

func (s *ClassService) AddStudent(ctx context.Context, req domain.AddStudentRequest, teacherID uuid.UUID) (*domain.User, string, error) {
	// 1. Verify Teacher owns the class
	class, err := s.GetClass(ctx, req.ClassID, teacherID)
	if err != nil {
		return nil, "", err
	}

	// 2. CHECK LIMITS: Validate limits on the Teacher's plan (always check before adding)
	teacher, err := s.userRepo.GetByID(ctx, teacherID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get teacher info: %w", err)
	}

	if teacher.StudentsLimit != -1 && teacher.StudentsCount >= teacher.StudentsLimit {
		return nil, "", errors.New("student limit reached. please upgrade your plan")
	}

	// 3. Check if student already exists by IIN
	existingUser, err := s.userRepo.GetByIIN(ctx, req.IIN)
	if err != nil {
		return nil, "", err
	}

	var studentID uuid.UUID
	var user *domain.User
	var password string

	if existingUser != nil {
		// Student exists, just add to class
		// Security check: ensure role is student
		if existingUser.Role != domain.RoleStudent {
			return nil, "", errors.New("user exists but is not a student")
		}

		// Check if student is ALREADY in this class to avoid double-counting/error
		inClass, err := s.repo.IsStudentInClass(ctx, class.ID, existingUser.ID)
		if err != nil {
			return nil, "", err
		}
		if inClass {
			return existingUser, "", nil // Already added, do nothing
		}

		studentID = existingUser.ID
		user = existingUser
	} else {
		// Create new student
		// Generate random 8-char password
		generatedPassword, err := utils.GenerateRandomString(8)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate password: %w", err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(generatedPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, "", fmt.Errorf("failed to hash password: %w", err)
		}

		var middleNamePtr *string
		if req.MiddleName != "" {
			middleNamePtr = &req.MiddleName
		}

		phone := utils.NormalizePhone(req.Phone)
		user = &domain.User{
			Phone:            &phone,
			IIN:              &req.IIN,
			PasswordHash:     string(hashedPassword),
			FirstName:        req.FirstName,
			LastName:         req.LastName,
			MiddleName:       middleNamePtr,
			BirthYear:        req.BirthYear,
			Gender:           domain.Gender(req.Gender),
			Role:             domain.RoleStudent,
			SchoolID:         class.SchoolID,
			SubscriptionPlan: domain.PlanFree,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", fmt.Errorf("failed to create student: %w", err)
		}

		studentID = user.ID
		password = generatedPassword
	}

	// 4. Add to class
	if err := s.repo.AddStudent(ctx, req.ClassID, studentID); err != nil {
		return nil, "", fmt.Errorf("failed to add student to class: %w", err)
	}

	// 5. Increment usage count for teacher
	if err := s.userRepo.IncrementStudentsCount(ctx, teacherID); err != nil {
		fmt.Printf("Warning: Failed to increment student count: %v\n", err)
	}

	// 6. Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_ADD_STUDENT",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", req.ClassID),
		Payload: map[string]interface{}{
			"student_id": studentID.String(),
			"is_new":     password != "",
		},
	})

	return user, password, nil
}

func (s *ClassService) DeleteClass(ctx context.Context, id int64, teacherID uuid.UUID) error {
	class, err := s.GetClass(ctx, id, teacherID)
	if err != nil {
		return err
	}

	// Capture count before deletion explicitly
	studentCount, err := s.repo.GetClassStudentCount(ctx, class.ID)
	if err != nil {
		log.Printf("Warning: Failed to fetch student count before deletion for class %d: %v\n", class.ID, err)
		studentCount = class.StudentCount // Fallback
	}
	log.Printf("[DEBUG] Deleting class %s (ID: %d). Releasing %d students for teacher %s\n", class.Name, class.ID, studentCount, teacherID)

	if err := s.repo.Delete(ctx, class.ID); err != nil {
		return err
	}

	// Update teacher limits (release slots)
	if studentCount > 0 {
		if err := s.userRepo.DecreaseStudentsCount(ctx, teacherID, int(studentCount)); err != nil {
			log.Printf("Warning: Failed to decrease student count after class deletion: %v\n", err)
		}
	} else {
		log.Printf("[DEBUG] No students to release for class %d\n", class.ID)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_DELETE",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
	})

	return nil
}

func (s *ClassService) RemoveStudent(ctx context.Context, classID int64, studentID uuid.UUID, teacherID uuid.UUID) error {
	class, err := s.GetClass(ctx, classID, teacherID)
	if err != nil {
		return err
	}

	if err := s.repo.RemoveStudent(ctx, class.ID, studentID); err != nil {
		return err
	}

	// Decrement usage count for teacher
	if err := s.userRepo.DecrementStudentsCount(ctx, teacherID); err != nil {
		fmt.Printf("Warning: Failed to decrement student count: %v\n", err)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_REMOVE_STUDENT",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
		Payload: map[string]interface{}{
			"student_id": studentID.String(),
		},
	})

	return nil
}

func (s *ClassService) GetClassStudents(ctx context.Context, classID int64, teacherID uuid.UUID) ([]domain.User, error) {
	class, err := s.GetClass(ctx, classID, teacherID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetStudents(ctx, class.ID)
}

func (s *ClassService) ListStudents(ctx context.Context, classID int64) ([]domain.User, error) {
	return s.repo.GetStudents(ctx, classID)
}

func (s *ClassService) CreateInvite(ctx context.Context, classID int64, teacherID uuid.UUID, req domain.CreateInviteRequest) (*domain.ClassInvite, error) {
	// 1. Verify Teacher owns the class
	class, err := s.GetClass(ctx, classID, teacherID)
	if err != nil {
		return nil, err
	}

	// 2. Generate uniq code
	code, err := utils.GenerateRandomString(12)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresInHours) * time.Hour)

	limit := req.UsageLimit
	if limit <= 0 {
		limit = 50 // Default fallback
	}

	invite := &domain.ClassInvite{
		ClassID:    class.ID,
		Code:       code,
		UsageLimit: limit,
		ExpiresAt:  expiresAt,
	}

	// 3. Expire all existing invites (as per requirement: "when new is created, old becomes invalid")
	if err := s.repo.ExpireAllInvites(ctx, invite.ClassID); err != nil {
		// Log error but continue? Or fail? Fail for strict consistency.
		return nil, err
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		return nil, err
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "CLASS_CREATE_INVITE",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
	})

	return invite, nil
}

func (s *ClassService) GetActiveInvite(ctx context.Context, classID int64, teacherID uuid.UUID) (*domain.ClassInvite, error) {
	// 1. Verify ownership (even though we just return the invite, teacher must own class)
	class, err := s.GetClass(ctx, classID, teacherID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetActiveInvite(ctx, class.ID)
}

func (s *ClassService) GetInviteInfo(ctx context.Context, code string) (*domain.InviteInfoResponse, error) {
	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if invite == nil {
		return nil, errors.New("invite not found")
	}

	// NEW: Check validity
	if time.Now().After(invite.ExpiresAt) {
		return nil, errors.New("invite expired")
	}
	if invite.UsageLimit > 0 && invite.UsageCount >= invite.UsageLimit {
		return nil, errors.New("invite limit reached")
	}

	class, err := s.repo.GetByID(ctx, invite.ClassID)
	if err != nil || class == nil {
		return nil, errors.New("class not found")
	}

	teacher, err := s.userRepo.GetByID(ctx, class.TeacherID)
	if err != nil || teacher == nil {
		return nil, errors.New("teacher not found")
	}

	return &domain.InviteInfoResponse{
		ClassName:   class.Name,
		TeacherName: teacher.FirstName + " " + teacher.LastName,
		ExpiresAt:   invite.ExpiresAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ValidateInviteAndLimits checks if invite is valid and teacher has space (publicly reachable via AuthService)
func (s *ClassService) ValidateInviteAndLimits(ctx context.Context, code string) error {
	// 1. Get and Validate Invite
	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		return err
	}
	if invite == nil {
		return errors.New("invite not found")
	}

	if time.Now().After(invite.ExpiresAt) {
		return errors.New("invite expired")
	}

	if invite.UsageLimit > 0 && invite.UsageCount >= invite.UsageLimit {
		return errors.New("invite limit reached")
	}

	// 2. Get Class Info
	class, err := s.repo.GetByID(ctx, invite.ClassID)
	if err != nil {
		return err
	}
	if class == nil {
		return errors.New("class not found")
	}

	// 3. CHECK TEACHER LIMITS
	teacher, err := s.userRepo.GetByID(ctx, class.TeacherID)
	if err != nil {
		return fmt.Errorf("failed to get teacher info: %w", err)
	}

	if teacher.StudentsLimit != -1 && teacher.StudentsCount >= teacher.StudentsLimit {
		return errors.New("teacher's student limit reached")
	}

	return nil
}

func (s *ClassService) AcceptInvite(ctx context.Context, code string, studentID uuid.UUID) error {
	// 1. Initial Validation
	if err := s.ValidateInviteAndLimits(ctx, code); err != nil {
		return err
	}

	// 2. Get Invite and Class again (for ID)
	invite, _ := s.repo.GetInviteByCode(ctx, code)
	class, _ := s.repo.GetByID(ctx, invite.ClassID)
	teacherID := class.TeacherID

	// 3. Check if student already in class
	inClass, err := s.repo.IsStudentInClass(ctx, class.ID, studentID)
	if err != nil {
		return err
	}
	if inClass {
		return nil // Already in class, just ignore
	}

	// 4. Add Student
	if err := s.repo.AddStudent(ctx, class.ID, studentID); err != nil {
		return err
	}

	// 5. Update Stats (Invite Usage + Teacher Student Count)
	if err := s.repo.IncrementInviteUsage(ctx, invite.ID); err != nil {
		fmt.Printf("Warning: Failed to increment invite usage for %d: %v\n", invite.ID, err)
	}

	if err := s.userRepo.IncrementStudentsCount(ctx, teacherID); err != nil {
		fmt.Printf("Warning: Failed to increment student count for teacher %s: %v\n", teacherID, err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &studentID,
		Role:     string(domain.RoleStudent),
		Action:   "CLASS_ACCEPT_INVITE",
		Entity:   "class",
		EntityID: fmt.Sprintf("%d", class.ID),
	})

	return nil
}

func (s *ClassService) GetStudentDashboard(ctx context.Context, studentID uuid.UUID) (*domain.StudentDashboardResponse, error) {
	// 1. Get student profile
	user, err := s.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("student not found")
	}

	// 2. Get student classes
	classes, err := s.repo.GetByStudentID(ctx, studentID)
	if err != nil {
		return nil, err
	}

	// 3. Transform to DTO
	studentClasses := make([]domain.StudentClass, 0)
	for _, c := range classes {
		teacher, _ := s.userRepo.GetByID(ctx, c.TeacherID)
		teacherName := "Unknown"
		if teacher != nil {
			teacherName = teacher.FirstName + " " + teacher.LastName
		}

		// Check if class has an active whiteboard session
		activeSession, _ := s.whiteboardRepo.GetActiveSessionByClass(ctx, c.ID)

		studentClasses = append(studentClasses, domain.StudentClass{
			ID:               c.ID,
			Name:             c.Name,
			TeacherName:      teacherName,
			CreatedAt:        c.CreatedAt.Format("2006-01-02 15:04:05"),
			HasActiveSession: activeSession != nil,
		})
	}

	return &domain.StudentDashboardResponse{
		User:    *user,
		Classes: studentClasses,
	}, nil
}
