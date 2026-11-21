package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/utils"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo         domain.UserRepository
	auditService *AuditService
}

func NewUserService(repo domain.UserRepository, audit *AuditService) *UserService {
	return &UserService{
		repo:         repo,
		auditService: audit,
	}
}

// GetProfile returns user profile by ID
func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.ProfileResponse, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Determine active plan and expiration
	activePlan := user.SubscriptionPlan
	expiresAt := user.SubscriptionExpiresAt
	isExpired := expiresAt != nil && expiresAt.Before(time.Now())

	if isExpired {
		activePlan = domain.PlanFree
		expiresAt = nil // Free plan has no expiration
	}

	// Calculate remaining limits based on active plan
	limits := domain.GetPlanLimits(activePlan)
	tokensRemaining := limits.TokensLimit - user.TokensUsed
	if activePlan == domain.PlanUltra {
		tokensRemaining = -1 // Unlimited
	}

	studentsRemaining := limits.StudentsLimit - user.StudentsCount
	if activePlan == domain.PlanUltra {
		studentsRemaining = -1 // Unlimited
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	iin := ""
	if user.IIN != nil {
		iin = *user.IIN
	}

	profile := &domain.ProfileResponse{
		ID:        user.ID.String(),
		Phone:     phone,
		Email:     email,
		IIN:       iin,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		BirthYear: user.BirthYear,
		Gender:    user.Gender,
		Role:      user.Role,
		SchoolID:  user.SchoolID,
		AvatarURL: user.AvatarURL,
		Subscription: domain.Subscription{
			Plan:              activePlan,
			TokensUsed:        user.TokensUsed,
			TokensLimit:       limits.TokensLimit,
			TokensRemaining:   tokensRemaining,
			StudentsCount:     user.StudentsCount,
			StudentsLimit:     limits.StudentsLimit,
			StudentsRemaining: studentsRemaining,
			StartedAt:         user.SubscriptionStartedAt,
			ExpiresAt:         expiresAt,
			LimitsResetAt:     new(time.Time),
		},
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.IIN != nil {
		profile.IIN = *user.IIN
	}

	*profile.Subscription.LimitsResetAt = domain.CalculateNextResetDate(user.SubscriptionStartedAt)

	return profile, nil
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req domain.UpdateProfileRequest) (*domain.ProfileResponse, error) {
	// Get current user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update only provided fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.BirthYear != nil {
		user.BirthYear = *req.BirthYear
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.SchoolID != nil {
		if user.Role != domain.RoleStudent {
			user.SchoolID = req.SchoolID
		}
	}
	if req.IIN != nil {
		user.IIN = req.IIN
	}

	// Save changes
	if err := s.repo.UpdateByID(ctx, userID, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_UPDATE_PROFILE",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	// Return updated profile
	return s.GetProfile(ctx, userID)
}

func (s *UserService) TeacherUpdateStudent(ctx context.Context, teacherID uuid.UUID, studentID uuid.UUID, req domain.UpdateProfileRequest) (*domain.ProfileResponse, error) {
	// 1. Get student
	student, err := s.repo.GetByID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.New("student not found")
	}

	if student.Role != domain.RoleStudent {
		return nil, errors.New("can only update student profiles")
	}

	// 2. Security Check: In real app, we should check if teacher "owns" this student
	// For now, allow any teacher (or we'd need to inject ClassRepo here)
	// Given Rule 19, if School mode, they must be in same school.

	// 3. Update fields
	if req.FirstName != nil {
		student.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		student.LastName = *req.LastName
	}
	if req.BirthYear != nil {
		student.BirthYear = *req.BirthYear
	}
	if req.Gender != nil {
		student.Gender = *req.Gender
	}

	if err := s.repo.UpdateByID(ctx, studentID, student); err != nil {
		return nil, err
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "TEACHER_UPDATE_STUDENT",
		Entity:   "user",
		EntityID: studentID.String(),
	})

	return s.GetProfile(ctx, studentID)
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID uuid.UUID, req domain.ChangePasswordRequest) error {
	// Get current user
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return fmt.Errorf("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.repo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_CHANGE_PASSWORD",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return nil
}

// UpdateEmail updates user email after verification
func (s *UserService) UpdateEmail(ctx context.Context, userID uuid.UUID, newEmail string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Email = &newEmail
	if err := s.repo.UpdateByID(ctx, userID, user); err != nil {
		return err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_UPDATE_EMAIL",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return nil
}

// UpdatePhone updates user phone after verification
func (s *UserService) UpdatePhone(ctx context.Context, userID uuid.UUID, newPhone string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	phone := utils.NormalizePhone(newPhone)
	user.Phone = &phone
	if err := s.repo.UpdateByID(ctx, userID, user); err != nil {
		return err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_UPDATE_PHONE",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return nil
}

// ResetStudentPassword generates a new password for a student (Teacher only)
func (s *UserService) ResetStudentPassword(ctx context.Context, teacherID uuid.UUID, studentID uuid.UUID) (string, error) {
	// 1. Get student
	student, err := s.repo.GetByID(ctx, studentID)
	if err != nil {
		return "", err
	}
	if student == nil {
		return "", errors.New("student not found")
	}
	if student.Role != domain.RoleStudent {
		return "", errors.New("cannot reset password for non-student users")
	}

	// 2. Generate new password
	newPassword, err := utils.GenerateRandomString(8)
	if err != nil {
		return "", err
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// 4. Update password in DB
	if err := s.repo.UpdatePassword(ctx, studentID, string(hashedPassword)); err != nil {
		return "", err
	}

	// 5. Log action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &teacherID,
		Role:     string(domain.RoleTeacher),
		Action:   "TEACHER_RESET_STUDENT_PASSWORD",
		Entity:   "user",
		EntityID: studentID.String(),
	})

	return newPassword, nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, fileData []byte, ext string) (string, error) {
	// 1. Create directory if not exists
	uploadDir := "./uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// 2. Generate filename
	filename := fmt.Sprintf("%s%s", userID.String(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// 3. Save file
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// 4. Update DB (store relative path or full URL)
	// For now, relative path that will be prefixed by API base URL in frontend or here
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)
	if err := s.repo.UpdateAvatar(ctx, userID, avatarURL); err != nil {
		return "", fmt.Errorf("failed to update avatar in DB: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &userID,
		Action:   "USER_UPDATE_AVATAR",
		Entity:   "user",
		EntityID: userID.String(),
	})

	return avatarURL, nil
}
func (s *UserService) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	// 1. Get current user to find filename
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. If avatar exists, try to delete file
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		// Extract filename from URL (e.g. /uploads/avatars/uuid.jpg)
		filename := filepath.Base(*user.AvatarURL)
		filePath := filepath.Join("./uploads/avatars", filename)
		
		// Ignore error if file doesn't exist
		_ = os.Remove(filePath)
	}

	// 3. Update DB
	if err := s.repo.UpdateAvatar(ctx, userID, ""); err != nil {
		return fmt.Errorf("failed to clear avatar in DB: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &userID,
		Action:   "USER_DELETE_AVATAR",
		Entity:   "user",
		EntityID: userID.String(),
	})

	return nil
}
