package handler

import (
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"io"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"zeinai-backend/internal/pkg/validation"
)

type UserHandler struct {
	service             *service.UserService
	verificationService *service.VerificationService
	validator           *validator.Validate
}

func NewUserHandler(s *service.UserService, v *service.VerificationService) *UserHandler {
	return &UserHandler{
		service:             s,
		verificationService: v,
		validator:           validation.NewValidator(),
	}
}

// GetProfile godoc
// @Summary Get current user profile
// @Description Get authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.ProfileResponse}
// @Failure 401 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /users/me [get]
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	profile, err := h.service.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	}

	return response.Success(c, profile)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user's profile information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpdateProfileRequest true "Profile Update"
// @Success 200 {object} response.APIResponse{data=domain.ProfileResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me [put]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	profile, err := h.service.UpdateProfile(c.Context(), userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, profile)
}

// ChangePassword godoc
// @Summary Change user password
// @Description Change authenticated user's password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.ChangePasswordRequest true "Password Change"
// @Success 200 {object} response.APIResponse{data=map[string]string}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me/password [put]
func (h *UserHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.ChangePassword(c.Context(), userID, req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "PASSWORD_CHANGE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Password changed successfully"})
}

// UpdateStudent godoc
// @Summary Update student profile
// @Description Update student profile information (Teacher only)
// @Tags Teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Param request body domain.UpdateProfileRequest true "Profile Update"
// @Success 200 {object} response.APIResponse{data=domain.ProfileResponse}
// @Router /teacher/students/{id} [put]
func (h *UserHandler) UpdateStudent(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid student ID format")
	}

	teacherID := c.Locals("user_id").(uuid.UUID)

	var req domain.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	profile, err := h.service.TeacherUpdateStudent(c.Context(), teacherID, studentID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, profile)
}

func (h *UserHandler) RequestEmailChange(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	err := h.verificationService.SendCode(c.Context(), "", req.Email, domain.VerificationTypeEmailChange)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SEND_FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Verification code sent to new email"})
}

func (h *UserHandler) VerifyEmailChange(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	var req struct {
		Email string `json:"email" validate:"required,email"`
		Code  string `json:"code" validate:"required,len=6"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	if err := h.verificationService.VerifyCode(c.Context(), "", req.Email, req.Code, domain.VerificationTypeEmailChange); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
	}
	if err := h.service.UpdateEmail(c.Context(), userID, req.Email); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Email updated successfully"})
}

func (h *UserHandler) RequestPhoneChange(c *fiber.Ctx) error {
	var req struct {
		Phone string `json:"phone" validate:"required,min=10,max=16"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	err := h.verificationService.SendCode(c.Context(), req.Phone, "", domain.VerificationTypePhoneChange)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "SEND_FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Verification code sent to new phone"})
}

func (h *UserHandler) VerifyPhoneChange(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	var req struct {
		Phone string `json:"phone" validate:"required,min=10,max=16"`
		Code  string `json:"code" validate:"required,len=6"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	if err := h.verificationService.VerifyCode(c.Context(), req.Phone, "", req.Code, domain.VerificationTypePhoneChange); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
	}
	if err := h.service.UpdatePhone(c.Context(), userID, req.Phone); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Phone updated successfully"})
}

// ResetStudentPassword godoc
// @Summary Reset student password
// @Description Reset student password and return the new one (Teacher only)
// @Tags Teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Student ID"
// @Success 200 {object} response.APIResponse{data=map[string]string}
// @Router /teacher/students/{id}/reset-password [put]
func (h *UserHandler) ResetStudentPassword(c *fiber.Ctx) error {
	studentID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid student ID format")
	}

	teacherID := c.Locals("user_id").(uuid.UUID)

	newPassword, err := h.service.ResetStudentPassword(c.Context(), teacherID, studentID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "RESET_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"password": newPassword,
	})
}

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload and update authenticated user's profile picture
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param avatar formData file true "Avatar image file"
// @Success 200 {object} response.APIResponse{data=map[string]string}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me/avatar [post]
func (h *UserHandler) UploadAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_FILE", "Failed to get file from request")
	}

	// Validate file size (e.g., 5MB)
	if file.Size > 5*1024*1024 {
		return response.Error(c, fiber.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds 5MB limit")
	}

	// Validate extension
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_FORMAT", "Only .jpg, .jpeg, .png, .webp are allowed")
	}

	src, err := file.Open()
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FILE_ERROR", "Failed to open file")
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FILE_ERROR", "Failed to read file data")
	}

	url, err := h.service.UpdateAvatar(c.Context(), userID, fileData, ext)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"avatar_url": url,
	})
}
// DeleteAvatar godoc
// @Summary Delete user avatar
// @Description Delete authenticated user's profile picture
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /users/me/avatar [delete]
func (h *UserHandler) DeleteAvatar(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	if err := h.service.DeleteAvatar(c.Context(), userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Avatar deleted successfully"})
}
