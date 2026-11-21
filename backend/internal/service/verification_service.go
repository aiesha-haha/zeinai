package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/mailer"
	"zeinai-backend/internal/pkg/mobizon"
	"zeinai-backend/internal/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	CodeLength     = 6
	CodeExpiration = 15 * time.Minute
	MaxAttempts    = 3
	RateLimitHour  = 5 // 5 requests per hour
)

type VerificationService struct {
	repo         domain.VerificationRepository
	userRepo     domain.UserRepository
	smsClient    *mobizon.MobizonClient
	emailClient  *mailer.Mailer
	auditService *AuditService
}

func NewVerificationService(
	repo domain.VerificationRepository,
	userRepo domain.UserRepository,
	smsClient *mobizon.MobizonClient,
	emailClient *mailer.Mailer,
	auditService *AuditService,
) *VerificationService {
	return &VerificationService{
		repo:         repo,
		userRepo:     userRepo,
		smsClient:    smsClient,
		emailClient:  emailClient,
		auditService: auditService,
	}
}

// SendCode sends a verification code via SMS or Email with progressive rate limiting
func (s *VerificationService) SendCode(ctx context.Context, phone, email string, codeType domain.VerificationType) error {
	identifier := email
	idType := "email"
	if identifier == "" {
		identifier = utils.NormalizePhone(phone)
		idType = "phone"
		phone = identifier
	}

	// 1. Check if blocked
	block, err := s.repo.GetBlock(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to check blocks: %w", err)
	}
	if block != nil {
		if block.IsPermanentlyBlocked {
			return fmt.Errorf("identifier is permanently blocked. Please contact support")
		}
		if block.BlockedUntil != nil && time.Now().Before(*block.BlockedUntil) {
			return fmt.Errorf("identifier is temporarily blocked until %s", block.BlockedUntil.Format("15:04:05"))
		}
	}

	// 2. Check progressive rate limit
	// Intervals: 1min, 3min, 5min, 1h, 24h
	totalSent, err := s.repo.GetIdentifierStats(ctx, identifier, 24*60) // look back 24h
	if err != nil {
		return fmt.Errorf("failed to check stats: %w", err)
	}

	lastCodeTime, err := s.repo.GetLastCodeTime(ctx, identifier)
	if err != nil {
		return fmt.Errorf("failed to check last code: %w", err)
	}

	if !lastCodeTime.IsZero() {
		var waitTime time.Duration
		switch totalSent {
		case 1:
			waitTime = 1 * time.Minute
		case 2:
			waitTime = 3 * time.Minute
		case 3:
			waitTime = 5 * time.Minute
		case 4:
			waitTime = 1 * time.Hour
		case 5:
			waitTime = 24 * time.Hour
			
			// Create explicit 24h block record so admin can see and remove it
			until := lastCodeTime.Add(waitTime)
			s.repo.CreateBlock(ctx, &domain.VerificationBlock{
				Identifier:           identifier,
				Type:                 idType,
				IsPermanentlyBlocked: false,
				BlockedUntil:         &until,
				Reason:               "Too many verification attempts (24h limit)",
			})
		default:
			if totalSent >= 6 {
				// Permanent block after 5 attempts in 24h
				s.repo.CreateBlock(ctx, &domain.VerificationBlock{
					Identifier:           identifier,
					Type:                 idType,
					IsPermanentlyBlocked: true,
					Reason:               "Too many verification attempts",
				})

				s.auditService.Log(ctx, &domain.AuditLog{
					Action:   "verification_blocked",
					Entity:   "verification",
					EntityID: identifier,
					Source:   "system",
					Role:     "student/teacher",
				})

				return fmt.Errorf("too many requests. Identifier has been permanently blocked")
			}
		}

		if time.Now().Before(lastCodeTime.Add(waitTime)) {
			nextTime := lastCodeTime.Add(waitTime)

			s.auditService.Log(ctx, &domain.AuditLog{
				Action:   "verification_rate_limit",
				Entity:   "verification",
				EntityID: identifier,
				Source:   "system",
				Role:     "student/teacher",
			})

			return fmt.Errorf("too many requests. Please try again after %s", nextTime.Format("15:04:05"))
		}
	}

	// Check user existence
	var user *domain.User
	if email != "" {
		user, _ = s.userRepo.GetByEmail(ctx, email)
	} else {
		user, _ = s.userRepo.GetByPhone(ctx, phone)
	}

	if (codeType == domain.VerificationTypePasswordReset) && user == nil {
		return fmt.Errorf("user not found")
	}
	if (codeType == domain.VerificationTypeRegistration || codeType == domain.VerificationTypeEmailChange || codeType == domain.VerificationTypePhoneChange) && user != nil {
		return fmt.Errorf("user already exists")
	}

	// Generate code
	code, err := generateCode(CodeLength)
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	// Create record
	vc := &domain.VerificationCode{
		Phone:      phone,
		Email:      email,
		Code:       code,
		Type:       codeType,
		IsVerified: false,
		Attempts:   0,
		ExpiresAt:  time.Now().Add(CodeExpiration),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, vc); err != nil {
		return fmt.Errorf("failed to save code: %w", err)
	}

	// Send
	if email != "" {
		if s.emailClient != nil {
			if codeType == domain.VerificationTypeRegistration {
				err = s.emailClient.SendVerificationCode(email, code)
			} else {
				err = s.emailClient.SendPasswordResetCode(email, code)
			}
			if err != nil {
				return fmt.Errorf("failed to send email: %w", err)
			}
		} else {
			fmt.Printf("📧 MOCK EMAIL to %s: Code %s\n", email, code)
		}
	} else {
		if s.smsClient != nil {
			if codeType == domain.VerificationTypeRegistration {
				err = s.smsClient.SendVerificationCode(phone, code)
			} else {
				err = s.smsClient.SendPasswordResetCode(phone, code)
			}
			if err != nil {
				return fmt.Errorf("failed to send SMS: %w", err)
			}
		} else {
			fmt.Printf("📱 MOCK SMS to %s: Code %s\n", phone, code)
		}
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		Action:   "verification_code_sent",
		Entity:   "verification",
		EntityID: identifier,
		Source:   "system",
		Role:     "student/teacher",
	})

	return nil
}

// VerifyCode verifies a code for phone or email
func (s *VerificationService) VerifyCode(ctx context.Context, phone, email, code string, codeType domain.VerificationType) error {
	var vc *domain.VerificationCode
	var err error

	if email != "" {
		vc, err = s.repo.GetByEmail(ctx, email, codeType)
	} else {
		phone = utils.NormalizePhone(phone)
		vc, err = s.repo.GetByPhone(ctx, phone, codeType)
	}

	if err != nil {
		return fmt.Errorf("failed to get verification code: %w", err)
	}
	if vc == nil {
		return fmt.Errorf("verification code not found or expired")
	}
	if vc.IsVerified {
		return fmt.Errorf("code already used")
	}
	if vc.Attempts >= MaxAttempts {
		return fmt.Errorf("too many failed attempts. Please request a new code")
	}
	if time.Now().After(vc.ExpiresAt) {
		return fmt.Errorf("verification code expired")
	}

	if vc.Code != code {
		id, _ := uuid.Parse(vc.ID)
		s.repo.IncrementAttempts(ctx, id)
		return fmt.Errorf("invalid verification code")
	}

	id, _ := uuid.Parse(vc.ID)
	if err := s.repo.Verify(ctx, id); err != nil {
		return fmt.Errorf("failed to verify code: %w", err)
	}

	return nil
}

// ResetPassword resets password via phone or email
func (s *VerificationService) ResetPassword(ctx context.Context, phone, email, code, newPassword string) error {
	if err := s.VerifyCode(ctx, phone, email, code, domain.VerificationTypePasswordReset); err != nil {
		return err
	}

	var user *domain.User
	var err error
	if email != "" {
		user, err = s.userRepo.GetByEmail(ctx, email)
	} else {
		phone = utils.NormalizePhone(phone)
		user, err = s.userRepo.GetByPhone(ctx, phone)
	}

	if err != nil || user == nil {
		return fmt.Errorf("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// GetStatus returns the current verification status for an identifier
func (s *VerificationService) GetStatus(ctx context.Context, identifier string) (*domain.VerificationStatusResponse, error) {
	if !utils.IsEmail(identifier) {
		identifier = utils.NormalizePhone(identifier)
	}

	resp := &domain.VerificationStatusResponse{
		CanSend: true,
	}

	// 1. Check if blocked
	block, err := s.repo.GetBlock(ctx, identifier)
	if err != nil {
		return nil, err
	}
	if block != nil {
		resp.Blocked = true
		resp.Reason = block.Reason
		resp.BlockedUntil = block.BlockedUntil
		if block.IsPermanentlyBlocked {
			resp.CanSend = false
			return resp, nil
		}
		if block.BlockedUntil != nil && time.Now().Before(*block.BlockedUntil) {
			resp.CanSend = false
			return resp, nil
		}
	}

	// 2. Check rate limit
	totalSent, err := s.repo.GetIdentifierStats(ctx, identifier, 24*60)
	if err != nil {
		return nil, err
	}

	lastCodeTime, err := s.repo.GetLastCodeTime(ctx, identifier)
	if err != nil {
		return nil, err
	}

	if !lastCodeTime.IsZero() {
		var waitTime time.Duration
		switch totalSent {
		case 1:
			waitTime = 1 * time.Minute
		case 2:
			waitTime = 3 * time.Minute
		case 3:
			waitTime = 5 * time.Minute
		case 4:
			waitTime = 1 * time.Hour
		case 5:
			waitTime = 24 * time.Hour
		default:
			if totalSent >= 6 {
				resp.Blocked = true
				resp.CanSend = false
				resp.Reason = "Too many attempts"
				return resp, nil
			}
		}

		nextAllowed := lastCodeTime.Add(waitTime)
		if time.Now().Before(nextAllowed) {
			resp.CanSend = false
			resp.NextSendAt = &nextAllowed
		}
	}

	return resp, nil
}

// GetAllBlocks returns all blocks for admin view
func (s *VerificationService) GetAllBlocks(ctx context.Context) ([]*domain.VerificationBlock, error) {
	return s.repo.GetAllBlocks(ctx)
}

// Unblock removes a block and clears history for an identifier (admin only)
func (s *VerificationService) Unblock(ctx context.Context, identifier string) error {
	if !utils.IsEmail(identifier) {
		identifier = utils.NormalizePhone(identifier)
	}
	
	// 1. Delete blocking record
	if err := s.repo.DeleteBlock(ctx, identifier); err != nil {
		return err
	}
	
	// 2. Clear verification history (codes) so rate limit statistics are reset
	return s.repo.DeleteRecentCodes(ctx, identifier)
}

func generateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}
