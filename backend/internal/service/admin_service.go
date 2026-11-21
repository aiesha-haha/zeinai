package service

import (
	"context"
	"fmt"
	"time"
	"zeinai-backend/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	userRepo     domain.UserRepository
	auditService *AuditService
	notifService *NotificationService
}

func NewAdminService(userRepo domain.UserRepository, audit *AuditService, notif *NotificationService) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		auditService: audit,
		notifService: notif,
	}
}

func (s *AdminService) ListTeachers(ctx context.Context, limit, offset int) ([]domain.User, error) {
	return s.userRepo.ListByRole(ctx, domain.RoleTeacher, limit, offset)
}

func (s *AdminService) GetTeacher(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Role != domain.RoleTeacher {
		return nil, fmt.Errorf("teacher not found")
	}
	return user, nil
}

func (s *AdminService) UpdateTeacher(ctx context.Context, id uuid.UUID, req domain.UpdateTeacherRequest, adminID uuid.UUID) (*domain.User, error) {
	user, err := s.GetTeacher(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.MiddleName != nil {
		user.MiddleName = req.MiddleName
	}
	if req.BirthYear != nil {
		user.BirthYear = *req.BirthYear
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Email != nil {
		// Treat empty string as "clear the field"
		if *req.Email == "" {
			user.Email = nil
		} else {
			// Check uniqueness: ensure no other user has this email
			existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
			}
			if existing != nil && existing.ID != id {
				return nil, fmt.Errorf("email already in use by another user")
			}
			user.Email = req.Email
		}
	}
	if req.Phone != nil {
		// Treat empty string as "clear the field"
		if *req.Phone == "" {
			user.Phone = nil
		} else {
			// Check uniqueness: ensure no other user has this phone
			existing, err := s.userRepo.GetByPhone(ctx, *req.Phone)
			if err != nil {
				return nil, fmt.Errorf("failed to check phone uniqueness: %w", err)
			}
			if existing != nil && existing.ID != id {
				return nil, fmt.Errorf("phone number already in use by another user")
			}
			user.Phone = req.Phone
		}
	}

	if err := s.userRepo.UpdateByID(ctx, id, user); err != nil {
		return nil, err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_UPDATE_TEACHER",
		Entity:   "user",
		EntityID: id.String(),
	})

	return user, nil
}

func (s *AdminService) DeleteTeacher(ctx context.Context, id uuid.UUID, adminID uuid.UUID) error {
	_, err := s.GetTeacher(ctx, id)
	if err != nil {
		return err
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_DELETE_TEACHER",
		Entity:   "user",
		EntityID: id.String(),
	})

	return nil
}

func (s *AdminService) UpdateSubscription(ctx context.Context, id uuid.UUID, req domain.UpdateSubscriptionRequest, adminID uuid.UUID) error {
	user, err := s.GetTeacher(ctx, id)
	if err != nil {
		return err
	}

	var expiresAt *time.Time
	if req.Plan != domain.PlanFree {
		if req.DurationMonths > 0 {
			// Calculate new expiration date
			start := time.Now()
			if user.SubscriptionExpiresAt != nil && user.SubscriptionExpiresAt.After(time.Now()) {
				start = *user.SubscriptionExpiresAt
			}
			expiry := start.AddDate(0, req.DurationMonths, 0)
			expiresAt = &expiry
		} else {
			// If DurationMonths is 0, preserve existing expiration
			expiresAt = user.SubscriptionExpiresAt
		}
	}

	if err := s.userRepo.UpdateSubscription(ctx, id, req.Plan, req.StudentsLimit, req.TokensLimit, expiresAt); err != nil {
		return err
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_UPDATE_SUBSCRIPTION",
		Entity:   "user",
		EntityID: id.String(),
		Payload:  req,
	})

	return nil
}

func (s *AdminService) NotifyTeacher(ctx context.Context, req domain.AdminNotifyRequest, adminID uuid.UUID) error {
	_, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return err
	}

	return s.notifService.CreateNotification(ctx, req.UserID, &adminID, "", req.Title, req.Message, domain.NotificationSystem, "")
}

func (s *AdminService) NotifyAllTeachers(ctx context.Context, req domain.AdminBulkNotifyRequest, adminID uuid.UUID) error {
	// 1. Get all teachers (handling pagination if needed, but for simplicity let's fetch first 1000)
	teachers, err := s.userRepo.ListByRole(ctx, domain.RoleTeacher, 1000, 0)
	if err != nil {
		return err
	}

	// 2. Iterate and notify
	for _, teacher := range teachers {
		_ = s.notifService.CreateNotification(ctx, teacher.ID, &adminID, "", req.Title, req.Message, domain.NotificationSystem, "")
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_NOTIFY_ALL_TEACHERS",
		Entity:   "notification",
		EntityID: "bulk",
		Payload:  req,
	})

	return nil
}

func (s *AdminService) CreateTeacher(ctx context.Context, req domain.CreateTeacherRequest, adminID uuid.UUID) (*domain.User, error) {
	var candidate *domain.User

	if req.Email != "" {
		u, err := s.userRepo.GetByEmailWithDeleted(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if u != nil {
			if u.DeletedAt == nil {
				return nil, fmt.Errorf("user with this email already exists")
			}
			candidate = u
		}
	}

	if req.Phone != "" {
		u, err := s.userRepo.GetByPhoneWithDeleted(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if u != nil {
			if u.DeletedAt == nil {
				return nil, fmt.Errorf("user with this phone already exists")
			}
			if candidate != nil && candidate.ID != u.ID {
				return nil, fmt.Errorf("conflict: phone belongs to a different deleted account")
			}
			candidate = u
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if candidate != nil {
		candidate.FirstName = req.FirstName
		candidate.LastName = req.LastName
		candidate.PasswordHash = string(hashedPassword)
		candidate.Role = domain.RoleTeacher
		candidate.BirthYear = req.BirthYear
		candidate.Gender = req.Gender
		
		if req.Email != "" {
			candidate.Email = &req.Email
		} else {
			candidate.Email = nil
		}
		
		if req.Phone != "" {
			candidate.Phone = &req.Phone
		} else {
			candidate.Phone = nil
		}

		if err := s.userRepo.Restore(ctx, candidate); err != nil {
			return nil, fmt.Errorf("failed to restore teacher: %w", err)
		}

		s.auditService.Log(ctx, &domain.AuditLog{
			UserID:   &adminID,
			Role:     string(domain.RoleAdmin),
			Action:   "ADMIN_RESTORE_TEACHER",
			Entity:   "user",
			EntityID: candidate.ID.String(),
		})

		return candidate, nil
	}

	user := &domain.User{
		ID:           uuid.New(),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleTeacher,
		BirthYear:    req.BirthYear,
		Gender:       req.Gender,
	}
	if req.Email != "" {
		user.Email = &req.Email
	}
	if req.Phone != "" {
		user.Phone = &req.Phone
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create teacher: %w", err)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_CREATE_TEACHER",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return user, nil
}

func (s *AdminService) SetTeacherPassword(ctx context.Context, id uuid.UUID, req domain.AdminSetPasswordRequest, adminID uuid.UUID) error {
	_, err := s.GetTeacher(ctx, id)
	if err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, id, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &adminID,
		Role:     string(domain.RoleAdmin),
		Action:   "ADMIN_SET_TEACHER_PASSWORD",
		Entity:   "user",
		EntityID: id.String(),
	})

	return nil
}
