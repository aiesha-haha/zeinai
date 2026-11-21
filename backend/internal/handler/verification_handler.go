package handler

import (
	"log"
	"strings"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
)

type VerificationHandler struct {
	service   *service.VerificationService
	validator *validator.Validate
}

func NewVerificationHandler(s *service.VerificationService) *VerificationHandler {
	return &VerificationHandler{
		service:   s,
		validator: validation.NewValidator(),
	}
}

// SendCode godoc
// @Summary Send verification code
// @Description Send SMS or Email verification code
// @Tags Verification
// @Accept json
// @Produce json
// @Param request body domain.SendVerificationCodeRequest true "Send Code Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 429 {object} response.APIResponse
// @Router /verification/send-code [post]
func (h *VerificationHandler) SendCode(c *fiber.Ctx) error {
	var req domain.SendVerificationCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if req.Phone == "" && req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Either phone or email is required")
	}

	err := h.service.SendCode(c.Context(), req.Phone, req.Email, req.Type)
	if err != nil {
		if err.Error() == "too many requests. Please try again later" {
			return response.Error(c, fiber.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", err.Error())
		}
		if err.Error() == "user already exists" || err.Error() == "user with this identifier already exists" {
			return response.Error(c, fiber.StatusConflict, "USER_EXISTS", "user already exists")
		}
		if err.Error() == "user not found" {
			return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
		}
		if strings.Contains(err.Error(), "failed to send SMS") {
			// The SMS provider rejected the message (often: empty balance). Tell the
			// user clearly so they can register with e-mail instead.
			log.Printf("[Verification] SMS sending failed: %v", err)
			return response.Error(c, fiber.StatusServiceUnavailable, "SMS_UNAVAILABLE", "SMS could not be sent. Please register with e-mail.")
		}
		return response.Error(c, fiber.StatusInternalServerError, "SEND_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message":            "Verification code sent successfully",
		"expires_in_minutes": 15,
	})
}

// VerifyCode godoc
// @Summary Verify code
// @Description Verify SMS or Email code
// @Tags Verification
// @Accept json
// @Produce json
// @Param request body domain.VerifyCodeRequest true "Verify Code Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /verification/verify-code [post]
func (h *VerificationHandler) VerifyCode(c *fiber.Ctx) error {
	var req domain.VerifyCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if req.Phone == "" && req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Either phone or email is required")
	}

	// Try registration first
	err := h.service.VerifyCode(c.Context(), req.Phone, req.Email, req.Code, domain.VerificationTypeRegistration)
	if err != nil {
		// Try password reset
		err = h.service.VerifyCode(c.Context(), req.Phone, req.Email, req.Code, domain.VerificationTypePasswordReset)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "VERIFICATION_FAILED", err.Error())
		}
	}

	return response.Success(c, fiber.Map{
		"message": "Code verified successfully",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password with verification code
// @Tags Verification
// @Accept json
// @Produce json
// @Param request body domain.ResetPasswordRequest true "Reset Password Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /verification/reset-password [post]
func (h *VerificationHandler) ResetPassword(c *fiber.Ctx) error {
	var req domain.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if req.Phone == "" && req.Email == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Either phone or email is required")
	}

	err := h.service.ResetPassword(c.Context(), req.Phone, req.Email, req.Code, req.NewPassword)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "RESET_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Password reset successfully",
	})
}

// GetStatus godoc
// @Summary Get verification status
// @Description Get current verification rate-limit status for phone or email
// @Tags Verification
// @Produce json
// @Param identifier query string true "Phone or Email"
// @Success 200 {object} domain.VerificationStatusResponse
// @Router /verification/status [get]
func (h *VerificationHandler) GetStatus(c *fiber.Ctx) error {
	identifier := c.Query("identifier")
	if identifier == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Identifier is required")
	}

	status, err := h.service.GetStatus(c.Context(), identifier)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "STATUS_CHECK_FAILED", err.Error())
	}

	return response.Success(c, status)
}

// GetAllBlocks godoc
// @Summary Get all blocks (Admin)
// @Description Get list of all blocked identifiers
// @Tags Verification
// @Security BearerAuth
// @Produce json
// @Success 200 {array} domain.VerificationBlock
// @Router /admin/verification/blocks [get]
func (h *VerificationHandler) GetAllBlocks(c *fiber.Ctx) error {
	blocks, err := h.service.GetAllBlocks(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "GET_BLOCKS_FAILED", err.Error())
	}

	return response.Success(c, blocks)
}

// Unblock godoc
// @Summary Unblock identifier (Admin)
// @Description Remove block for phone or email
// @Tags Verification
// @Security BearerAuth
// @Produce json
// @Param identifier query string true "Phone or Email to unblock"
// @Success 200 {object} response.APIResponse
// @Router /admin/verification/unblock [delete]
func (h *VerificationHandler) Unblock(c *fiber.Ctx) error {
	identifier := c.Query("identifier")
	if identifier == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "Identifier is required")
	}

	err := h.service.Unblock(c.Context(), identifier)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UNBLOCK_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Identifier unblocked successfully",
	})
}
