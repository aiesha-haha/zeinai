package handler

import (
	"fmt"
	"time"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/oauth"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"
	"golang.org/x/oauth2"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	service   *service.AuthService
	validator *validator.Validate
}

func NewAuthHandler(s *service.AuthService) *AuthHandler {
	return &AuthHandler{
		service:   s,
		validator: validation.NewValidator(),
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with phone or email, iin, and school info
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "Register Request"
// @Success 201 {object} response.APIResponse{data=domain.LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req domain.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	if req.Phone == "" && req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Either phone or email is required")
	}

	// Public self-registration may only create students or teachers.
	// Admin / school accounts are provisioned by an administrator.
	if req.Role != string(domain.RoleStudent) && req.Role != string(domain.RoleTeacher) {
		return response.Error(c, fiber.StatusForbidden, "ROLE_NOT_ALLOWED", "Only student or teacher roles can be self-registered")
	}

	user, err := h.service.Register(c.Context(), req)
	if err != nil {
		if err.Error() == "user with this phone already exists" ||
			err.Error() == "user with this email already exists" ||
			err.Error() == "user with this IIN already exists" {
			return response.Error(c, fiber.StatusConflict, "USER_EXISTS", err.Error())
		}
		return response.Error(c, fiber.StatusConflict, "REGISTRATION_FAILED", err.Error())
	}

	return response.Created(c, user)
}

// Login godoc
// @Summary Login user
// @Description Login with phone/email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login Request"
// @Success 200 {object} response.APIResponse{data=domain.LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req domain.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	if req.Phone == "" && req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Either phone or email is required")
	}

	resp, err := h.service.Login(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "LOGIN_FAILED", err.Error())
	}

	return response.Success(c, resp)
}

// LoginStudent godoc
// @Summary Login student
// @Description Login student with IIN and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.StudentLoginRequest true "Student Login Request"
// @Success 200 {object} response.APIResponse{data=domain.LoginResponse}
// @Failure 401 {object} response.APIResponse
// @Router /auth/login/student [post]
func (h *AuthHandler) LoginStudent(c *fiber.Ctx) error {
	var req domain.StudentLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	resp, err := h.service.LoginStudent(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "LOGIN_FAILED", err.Error())
	}

	return response.Success(c, resp)
}

// CheckIIN godoc
// @Summary Check if IIN exists
// @Description Check if a user with this IIN already exists (public)
// @Tags Auth
// @Param iin query string true "User IIN"
// @Success 200 {object} response.APIResponse{data=domain.CheckIINResponse}
// @Failure 400 {object} response.APIResponse
// @Router /auth/check-iin [get]
func (h *AuthHandler) CheckIIN(c *fiber.Ctx) error {
	iin := c.Query("iin")
	if len(iin) != 12 {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_IIN", "IIN must be 12 digits")
	}

	exists, user, err := h.service.CheckIIN(c.Context(), iin)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CHECK_IIN_FAILED", err.Error())
	}

	resp := map[string]interface{}{
		"exists": exists,
	}
	if exists && user != nil {
		// Only return minimal info to verify it's them
		resp["first_name"] = user.FirstName
		resp["last_name"] = user.LastName
	}

	return response.Success(c, resp)
}

// InitStudentReset godoc
// @Summary Init student password reset
// @Description Get masked phone number by IIN
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.StudentResetInitRequest true "Init Request"
// @Success 200 {object} response.APIResponse{data=domain.StudentResetInitResponse}
// @Failure 400 {object} response.APIResponse
// @Router /auth/reset-password/student/init [post]
func (h *AuthHandler) InitStudentReset(c *fiber.Ctx) error {
	var req domain.StudentResetInitRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	resp, err := h.service.InitStudentReset(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INIT_FAILED", err.Error())
	}

	return response.Success(c, resp)
}

// ConfirmStudentReset godoc
// @Summary Confirm student password reset
// @Description Reset password by entering hidden digits
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.StudentResetConfirmRequest true "Confirm Request"
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /auth/reset-password/student/confirm [post]
func (h *AuthHandler) ConfirmStudentReset(c *fiber.Ctx) error {
	var req domain.StudentResetConfirmRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	if err := h.service.ConfirmStudentReset(c.Context(), req); err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "RESET_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Password reset successfully"})
}

// Logout godoc
// @Summary Logout user
// @Description Logout current user and log the action
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	role, _ := c.Locals("role").(domain.Role)

	if err := h.service.Logout(c.Context(), userID, role); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "LOGOUT_FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Logged out successfully"})
}

// GoogleLogin godoc
// @Summary Redirect to Google Login
// @Description Redirect user to Google OAuth page
// @Tags Auth
// @Param action query string false "Action: login or register"
// @Router /auth/google/login [get]
func (h *AuthHandler) GoogleLogin(c *fiber.Ctx) error {
	action := c.Query("action", "login") // default to login
	state := fmt.Sprintf("%s:%d", action, time.Now().Unix())
	url := oauth.GetGoogleOAuthConfig().AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	return c.Redirect(url)
}

// GoogleCallback godoc
// @Summary Google OAuth Callback
// @Description Handle Google OAuth callback
// @Tags Auth
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_CODE", "Missing auth code")
	}

	// State parameter is validated but action is not needed for now
	_ = state

	googleUser, err := oauth.GetGoogleUserInfo(code)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "OAUTH_FAILED", err.Error())
	}

	// Try to login the user
	resp, err := h.service.LoginGoogle(c.Context(), googleUser)
	
	frontendURL := "https://zeinai.kz" // Production frontend URL
	
	if err != nil {
		if err.Error() == "GOOGLE_USER_NOT_REGISTERED" {
			// User doesn't exist - redirect to complete profile
			return c.Redirect(fmt.Sprintf("%s/google-complete-profile?google_id=%s&email=%s&first_name=%s&last_name=%s&picture=%s",
				frontendURL, googleUser.ID, googleUser.Email, googleUser.GivenName, googleUser.FamilyName, googleUser.Picture))
		}
		return c.Redirect(fmt.Sprintf("%s/login?error=AUTH_FAILED", frontendURL))
	}

	// User exists - redirect to login with tokens
	return c.Redirect(fmt.Sprintf("%s/login?token=%s&refresh=%s&role=%s", frontendURL, resp.AccessToken, resp.RefreshToken, resp.User.Role))
}

// GoogleRegisterComplete godoc
// @Summary Complete Google OAuth Registration
// @Description Complete registration for new Google users with required profile info
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body domain.GoogleRegisterCompleteRequest true "Complete Registration Request"
// @Success 200 {object} response.APIResponse{data=domain.LoginResponse}
// @Failure 400 {object} response.APIResponse
// @Router /auth/google/register/complete [post]
func (h *AuthHandler) GoogleRegisterComplete(c *fiber.Ctx) error {
	var req domain.GoogleRegisterCompleteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	resp, err := h.service.RegisterGoogle(c.Context(), req)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "REGISTRATION_FAILED", err.Error())
	}

	return response.Success(c, resp)
}

// Refresh godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body map[string]string true "Refresh Request"
// @Success 200 {object} response.APIResponse{data=domain.LoginResponse}
// @Failure 401 {object} response.APIResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	resp, err := h.service.RefreshToken(c.Context(), req.RefreshToken)
	if err != nil {
		return response.Error(c, fiber.StatusUnauthorized, "REFRESH_FAILED", err.Error())
	}

	return response.Success(c, resp)
}
