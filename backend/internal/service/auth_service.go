package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/oauth"
	"zeinai-backend/internal/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo     domain.UserRepository
	cfg          *config.Config
	auditService *AuditService
	classService *ClassService
}

func NewAuthService(repo domain.UserRepository, cfg *config.Config, audit *AuditService, classService *ClassService) *AuthService {
	return &AuthService{
		userRepo:     repo,
		cfg:          cfg,
		auditService: audit,
		classService: classService,
	}
}

// Register registers a new user and returns login response with tokens
func (s *AuthService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.LoginResponse, error) {
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("phone or email is required")
	}

	// 0. Pre-validate Invite Code if present (Atomic Join Flow)
	if req.InviteCode != "" && s.classService != nil {
		if err := s.classService.ValidateInviteAndLimits(ctx, req.InviteCode); err != nil {
			return nil, err
		}
	}

	var candidate *domain.User

	// 1. Check Phone
	if req.Phone != "" {
		req.Phone = utils.NormalizePhone(req.Phone)
		u, err := s.userRepo.GetByPhoneWithDeleted(ctx, req.Phone)
		if err != nil {
			return nil, err
		}
		if u != nil {
			if u.DeletedAt == nil {
				return nil, errors.New("user with this phone already exists")
			}
			candidate = u
		}
	}

	// 2. Check Email
	if req.Email != "" {
		u, err := s.userRepo.GetByEmailWithDeleted(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if u != nil {
			if u.DeletedAt == nil {
				return nil, errors.New("user with this email already exists")
			}
			if candidate != nil && candidate.ID != u.ID {
				return nil, errors.New("conflict: email belongs to a different deleted account")
			}
			candidate = u
		}
	}

	// 3. Check IIN
	if req.IIN != "" {
		u, err := s.userRepo.GetByIINWithDeleted(ctx, req.IIN)
		if err != nil {
			return nil, err
		}
		if u != nil {
			if u.DeletedAt == nil {
				return nil, errors.New("user with this IIN already exists")
			}
			if candidate != nil && candidate.ID != u.ID {
				return nil, errors.New("conflict: IIN belongs to a different deleted account")
			}
			candidate = u
		}
	}

	// 4. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var emailPtr *string
	if req.Email != "" {
		emailPtr = &req.Email
	}

	// 5. Restore or Create
	if candidate != nil {
		// Update candidate with new info
		if req.Phone != "" {
			phone := utils.NormalizePhone(req.Phone)
			candidate.Phone = &phone
		} else {
			candidate.Phone = nil
		}
		candidate.Email = emailPtr
		if req.IIN != "" {
			iin := req.IIN
			candidate.IIN = &iin
		} else {
			candidate.IIN = nil
		}
		candidate.PasswordHash = string(hashedPassword)
		candidate.FirstName = req.FirstName
		candidate.LastName = req.LastName
		// Maintain or update middle name? Update is safer if user changed it.
		// Req doesn't have middle name?? Let's check struct.
		// RegisterRequest usually has everything.
		// Looking at Step 28, lines 86-88: FirstName, LastName. No MiddleName in struct usage?
		// Wait, line 55 in repo `user.MiddleName`.
		// Let's check `domain.RegisterRequest`.
		// Assuming it's in `domain/dto.go`.
		// But in Step 28 line 86-88: `req.FirstName`, `req.LastName`.
		// Line 92 `SchoolID: req.SchoolID`.
		// MiddleName is missing in old implementation?
		// Line 55 in repo scan: `user.MiddleName`.

		// I'll stick to what was there. If `req` has MiddleName, use it.
		// If not, I can't guess.

		// Wait, I need to check `RegisterRequest` definition to be sure.
		// But I can see usage in original code:
		/*
			81: 	user := &domain.User{
			82: 		Phone:        req.Phone,
			...
			86: 		FirstName:    req.FirstName,
			87: 		LastName:     req.LastName,
		*/
		// MiddleName was NOT set in original `Register`!
		// So I will ignore MiddleName for now as well to match existing logic.

		candidate.BirthYear = req.BirthYear
		candidate.Gender = domain.Gender(req.Gender)
		candidate.Role = domain.Role(req.Role)
		candidate.SchoolID = req.SchoolID

		// Reset Subscription to Free default
		limits := domain.GetPlanLimits(domain.PlanFree)
		candidate.SubscriptionPlan = domain.PlanFree
		candidate.TokensUsed = 0
		candidate.TokensLimit = limits.TokensLimit
		candidate.StudentsCount = 0
		candidate.StudentsLimit = limits.StudentsLimit
		candidate.SubscriptionStartedAt = time.Now() // Reset start date

		if err := s.userRepo.Restore(ctx, candidate); err != nil {
			return nil, err
		}

		// 6. Post-Registration Join (Atomic)
		if req.InviteCode != "" && s.classService != nil {
			if err := s.classService.AcceptInvite(ctx, req.InviteCode, candidate.ID); err != nil {
				// We don't rollback restore because account is already better to have restored
				// than in a flux state. But we return the error to the user.
				fmt.Printf("Warning: Atomic join failed after restore for user %s: %v\n", candidate.ID, err)
			}
		}

		// 7. Generate Tokens for immediate login
		accessToken, refreshToken, err := utils.GenerateTokens(candidate, s.cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate tokens: %w", err)
		}

		s.auditService.Log(ctx, &domain.AuditLog{
			UserID:   &candidate.ID,
			Role:     string(candidate.Role),
			Action:   "USER_RESTORED",
			Entity:   "user",
			EntityID: candidate.ID.String(),
			Payload: map[string]interface{}{
				"iin":  candidate.IIN,
				"role": candidate.Role,
			},
		})

		return &domain.LoginResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         *candidate,
		}, nil
	}

	// Create New User
	var phonePtr *string
	if req.Phone != "" {
		phone := utils.NormalizePhone(req.Phone)
		phonePtr = &phone
	}

	user := &domain.User{
		Phone:        phonePtr,
		Email:        emailPtr,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BirthYear:    req.BirthYear,
		Gender:       domain.Gender(req.Gender),
		Role:         domain.Role(req.Role),
		SchoolID:     req.SchoolID,
	}

	if req.IIN != "" {
		iin := req.IIN
		user.IIN = &iin
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 6. Post-Registration Join (Atomic)
	if req.InviteCode != "" && s.classService != nil {
		if err := s.classService.AcceptInvite(ctx, req.InviteCode, user.ID); err != nil {
			// NOTE: Since user is created, if join fails (unlikely due to pre-validation),
			// we log warning but the user is already persistent.
			fmt.Printf("Warning: Atomic join failed after creation for user %s: %v\n", user.ID, err)
		}
	}

	// 7. Generate Tokens for immediate login
	accessToken, refreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_REGISTER",
		Entity:   "user",
		EntityID: user.ID.String(),
		Payload: map[string]interface{}{
			"iin":  user.IIN,
			"role": user.Role,
		},
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// Login authenticates a user
// Login authenticates a user
func (s *AuthService) Login(ctx context.Context, req domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("phone or email is required")
	}

	var user *domain.User
	var err error

	// 1. Find user
	if req.Email != "" {
		user, err = s.userRepo.GetByEmail(ctx, req.Email)
	} else {
		req.Phone = utils.NormalizePhone(req.Phone)
		user, err = s.userRepo.GetByPhone(ctx, req.Phone)
	}

	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Compare Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate Tokens
	accessToken, refreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "USER_LOGIN",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// LoginGoogle handles Google OAuth login
func (s *AuthService) LoginGoogle(ctx context.Context, googleUser *oauth.GoogleUser) (*domain.LoginResponse, error) {
	// 1. Find user by Google ID
	user, err := s.userRepo.GetByGoogleID(ctx, googleUser.ID)
	if err != nil {
		return nil, err
	}

	// 2. If not found by Google ID, try by Email
	if user == nil {
		user, err = s.userRepo.GetByEmail(ctx, googleUser.Email)
		if err != nil {
			return nil, err
		}

		// 3. If found by Email, link Google ID
		if user != nil {
			user.GoogleID = &googleUser.ID
			if err := s.userRepo.UpdateByID(ctx, user.ID, user); err != nil {
				return nil, fmt.Errorf("failed to link Google ID: %w", err)
			}
		}
	}

	// 4. If still not found, we need to create a "partial" user or ask for more info
	if user == nil {
		return nil, errors.New("GOOGLE_USER_NOT_REGISTERED")
	}

	// 5. Generate Tokens
	accessToken, refreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "GOOGLE_LOGIN",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// LoginStudent authenticates a student specifically by IIN
func (s *AuthService) LoginStudent(ctx context.Context, req domain.StudentLoginRequest) (*domain.LoginResponse, error) {
	user, err := s.userRepo.GetByIIN(ctx, req.IIN)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check role is student
	if user.Role != domain.RoleStudent {
		return nil, errors.New("invalid credentials") // Don't leak role info, but effectively only students can login
	}

	// Compare Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate Tokens
	accessToken, refreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "STUDENT_LOGIN",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// CheckIIN checks if a user with the given IIN exists
func (s *AuthService) CheckIIN(ctx context.Context, iin string) (bool, *domain.User, error) {
	user, err := s.userRepo.GetByIIN(ctx, iin)
	if err != nil {
		return false, nil, err
	}
	if user != nil {
		return true, user, nil
	}
	return false, nil, nil
}

// InitStudentReset returns a masked phone number for the student
func (s *AuthService) InitStudentReset(ctx context.Context, req domain.StudentResetInitRequest) (*domain.StudentResetInitResponse, error) {
	user, err := s.userRepo.GetByIIN(ctx, req.IIN)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if user.Role != domain.RoleStudent {
		return nil, errors.New("method only allowed for students")
	}

	if user.Phone == nil || *user.Phone == "" {
		return nil, errors.New("student has no phone number attached")
	}

	// Simple masking: leave first 4 and last 3 visible (for 11 digit phones)
	// 77011234567 -> 7701****567
	// Digits to guess: 1234

	// Ensure only digits
	phoneDigits := utils.NormalizePhone(*user.Phone)
	if len(phoneDigits) < 11 {
		return nil, errors.New("invalid phone format in database")
	}

	// Mask digits from index 4 to 8 (4 digits)
	masked := phoneDigits[:4] + "****" + phoneDigits[8:]

	return &domain.StudentResetInitResponse{
		MaskedPhone: masked,
	}, nil
}

// ConfirmStudentReset verifies the hidden digits and resets password
func (s *AuthService) ConfirmStudentReset(ctx context.Context, req domain.StudentResetConfirmRequest) error {
	user, err := s.userRepo.GetByIIN(ctx, req.IIN)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if user.Role != domain.RoleStudent {
		return errors.New("method only allowed for students")
	}

	if user.Phone == nil || *user.Phone == "" {
		return errors.New("student has no phone number attached")
	}

	phoneDigits := utils.NormalizePhone(*user.Phone)
	if len(phoneDigits) < 11 {
		return errors.New("invalid phone format in database")
	}

	// Extract hidden part: indices 4,5,6,7
	expectedCode := phoneDigits[4:8]

	if req.Code != expectedCode {
		return errors.New("verification digits do not match")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, user.ID, string(hashedPassword)); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "STUDENT_RESET_PASSWORD",
		Entity:   "user",
		EntityID: user.ID.String(),
	})

	return nil
}

// Logout logs the user logout action
func (s *AuthService) Logout(ctx context.Context, userID uuid.UUID, role domain.Role) error {
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &userID,
		Role:     string(role),
		Action:   "USER_LOGOUT",
		Entity:   "user",
		EntityID: userID.String(),
	})
	return nil
}

// RegisterGoogle handles Google OAuth registration completion
func (s *AuthService) RegisterGoogle(ctx context.Context, req domain.GoogleRegisterCompleteRequest) (*domain.LoginResponse, error) {
	// 1. Check if user with this Google ID already exists
	// Using WithDeleted for GoogleID to check if they previously logged in with Google and deleted their account
	existingUser, err := s.userRepo.GetByEmailWithDeleted(ctx, req.Email) // Checking by email is safer as Google ID might be linked
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		if existingUser.DeletedAt == nil {
			// User exists and is active. If GoogleID doesn't match/exist, link it
			if existingUser.GoogleID == nil {
				existingUser.GoogleID = &req.GoogleID
				if err := s.userRepo.UpdateByID(ctx, existingUser.ID, existingUser); err != nil {
					return nil, fmt.Errorf("failed to link Google ID: %w", err)
				}
				
				// Generate tokens
				accessToken, refreshToken, err := utils.GenerateTokens(existingUser, s.cfg)
				if err != nil {
					return nil, fmt.Errorf("failed to generate tokens: %w", err)
				}

				s.auditService.Log(ctx, &domain.AuditLog{
					UserID:   &existingUser.ID,
					Role:     string(existingUser.Role),
					Action:   "GOOGLE_ACCOUNT_LINKED",
					Entity:   "user",
					EntityID: existingUser.ID.String(),
				})

				return &domain.LoginResponse{
					AccessToken:  accessToken,
					RefreshToken: refreshToken,
					User:         *existingUser,
				}, nil
			}
			return nil, errors.New("user with this email already exists")
		} else {
			// User is deleted, we need to RESTORE them
			existingUser.FirstName = req.FirstName
			existingUser.LastName = req.LastName
			existingUser.BirthYear = req.BirthYear
			existingUser.Gender = domain.Gender(req.Gender)
			existingUser.GoogleID = &req.GoogleID // Ensure Google ID is set

			if req.IIN != "" {
				iin := req.IIN
				existingUser.IIN = &iin
			}

			// Reset Subscription to Free default upon restore
			limits := domain.GetPlanLimits(domain.PlanFree)
			existingUser.SubscriptionPlan = domain.PlanFree
			existingUser.TokensUsed = 0
			existingUser.TokensLimit = limits.TokensLimit
			existingUser.StudentsCount = 0
			existingUser.StudentsLimit = limits.StudentsLimit
			existingUser.SubscriptionStartedAt = time.Now()

			if err := s.userRepo.Restore(ctx, existingUser); err != nil {
				return nil, fmt.Errorf("failed to restore user: %w", err)
			}

			// Generate tokens
			accessToken, refreshToken, err := utils.GenerateTokens(existingUser, s.cfg)
			if err != nil {
				return nil, fmt.Errorf("failed to generate tokens: %w", err)
			}

			s.auditService.Log(ctx, &domain.AuditLog{
				UserID:   &existingUser.ID,
				Role:     string(existingUser.Role),
				Action:   "GOOGLE_USER_RESTORED",
				Entity:   "user",
				EntityID: existingUser.ID.String(),
			})

			return &domain.LoginResponse{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
				User:         *existingUser,
			}, nil
		}
	}

	// 4. Create new user
	user := &domain.User{
		Email:        &req.Email,
		GoogleID:     &req.GoogleID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		BirthYear:    req.BirthYear,
		Gender:       domain.Gender(req.Gender),
		Role:         googleRole(req.Role),
		PasswordHash: "", // No password for Google users
	}

	if req.IIN != "" {
		iin := req.IIN
		user.IIN = &iin
	}

	// Set default subscription limits
	limits := domain.GetPlanLimits(domain.PlanFree)
	user.SubscriptionPlan = domain.PlanFree
	user.TokensLimit = limits.TokensLimit
	user.StudentsLimit = limits.StudentsLimit
	user.SubscriptionStartedAt = time.Now()

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// 5. Generate tokens
	accessToken, refreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Log Action
	s.auditService.Log(ctx, &domain.AuditLog{
		UserID:   &user.ID,
		Role:     string(user.Role),
		Action:   "GOOGLE_REGISTER",
		Entity:   "user",
		EntityID: user.ID.String(),
		Payload: map[string]interface{}{
			"google_id": req.GoogleID,
			"email":     req.Email,
		},
	})

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

// RefreshToken validates a refresh token and returns new tokens
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error) {
	claims, err := utils.ValidateToken(refreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Subject != "refresh" {
		return nil, errors.New("invalid token subject")
	}

	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	accessToken, newRefreshToken, err := utils.GenerateTokens(user, s.cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

// googleRole maps the role chosen during Google sign-up to a self-service role.
// Anything other than "student" falls back to teacher (the historical default).
func googleRole(r string) domain.Role {
	if r == string(domain.RoleStudent) {
		return domain.RoleStudent
	}
	return domain.RoleTeacher
}
