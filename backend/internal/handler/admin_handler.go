package handler

import (
	"strconv"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AdminHandler struct {
	service   *service.AdminService
	validator *validator.Validate
}

func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{
		service:   s,
		validator: validation.NewValidator(),
	}
}

// ListTeachers godoc
// @Summary List teachers
// @Description Get all teachers with pagination (Admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} response.APIResponse{data=[]domain.User}
// @Router /admin/teachers [get]
func (h *AdminHandler) ListTeachers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	teachers, err := h.service.ListTeachers(c.Context(), limit, offset)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	return response.Success(c, teachers)
}

// GetTeacher godoc
// @Summary Get teacher details
// @Description Get specific teacher details (Admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Teacher ID"
// @Success 200 {object} response.APIResponse{data=domain.User}
// @Router /admin/teachers/{id} [get]
func (h *AdminHandler) GetTeacher(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid teacher ID")
	}

	teacher, err := h.service.GetTeacher(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", err.Error())
	}

	return response.Success(c, teacher)
}

// UpdateTeacher godoc
// @Summary Update teacher profile
// @Description Update teacher profile by admin
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Teacher ID"
// @Param request body domain.UpdateTeacherRequest true "Teacher Update"
// @Success 200 {object} response.APIResponse{data=domain.User}
// @Router /admin/teachers/{id} [put]
func (h *AdminHandler) UpdateTeacher(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid teacher ID")
	}

	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.UpdateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	teacher, err := h.service.UpdateTeacher(c.Context(), id, req, adminID)
	if err != nil {
		if err.Error() == "phone number already in use by another user" ||
			err.Error() == "email already in use by another user" {
			return response.Error(c, fiber.StatusConflict, "DUPLICATE_CONTACT", err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, teacher)
}

// DeleteTeacher godoc
// @Summary Delete teacher
// @Description Soft delete a teacher (Admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Teacher ID"
// @Success 200 {object} response.APIResponse
// @Router /admin/teachers/{id} [delete]
func (h *AdminHandler) DeleteTeacher(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid teacher ID")
	}

	adminID := c.Locals("user_id").(uuid.UUID)

	if err := h.service.DeleteTeacher(c.Context(), id, adminID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// UpdateSubscription godoc
// @Summary Update teacher subscription
// @Description Update teacher's plan and limits (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Teacher ID"
// @Param request body domain.UpdateSubscriptionRequest true "Subscription Update"
// @Success 200 {object} response.APIResponse
// @Router /admin/teachers/{id}/subscription [put]
func (h *AdminHandler) UpdateSubscription(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid teacher ID")
	}

	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.UpdateSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.UpdateSubscription(c.Context(), id, req, adminID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// NotifyTeacher godoc
// @Summary Send notification to teacher
// @Description Send a system notification to a specific teacher
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.AdminNotifyRequest true "Notification Data"
// @Success 200 {object} response.APIResponse
// @Router /admin/notify [post]
func (h *AdminHandler) NotifyTeacher(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.AdminNotifyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.NotifyTeacher(c.Context(), req, adminID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "NOTIFY_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// NotifyAllTeachers godoc
// @Summary Send notification to all teachers
// @Description Send a system notification to all teachers (bulk)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.AdminBulkNotifyRequest true "Bulk Notification Data"
// @Success 200 {object} response.APIResponse
// @Router /admin/notify-all [post]
func (h *AdminHandler) NotifyAllTeachers(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.AdminBulkNotifyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.NotifyAllTeachers(c.Context(), req, adminID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "NOTIFY_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// CreateTeacher godoc
// @Summary Create teacher
// @Description Create a new teacher account (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.CreateTeacherRequest true "Teacher Creation"
// @Success 201 {object} response.APIResponse{data=domain.User}
// @Router /admin/teachers [post]
func (h *AdminHandler) CreateTeacher(c *fiber.Ctx) error {
	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.CreateTeacherRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	teacher, err := h.service.CreateTeacher(c.Context(), req, adminID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"data":    teacher,
	})
}

// SetTeacherPassword godoc
// @Summary Set teacher password
// @Description Set a new password for a teacher (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Teacher ID"
// @Param request body domain.AdminSetPasswordRequest true "Password Data"
// @Success 200 {object} response.APIResponse
// @Router /admin/teachers/{id}/set-password [post]
func (h *AdminHandler) SetTeacherPassword(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid teacher ID")
	}

	adminID := c.Locals("user_id").(uuid.UUID)

	var req domain.AdminSetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.SetTeacherPassword(c.Context(), id, req, adminID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "SET_PASSWORD_FAILED", err.Error())
	}

	return response.Success(c, nil)
}
