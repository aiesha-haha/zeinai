package handler

import (
	"fmt"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type STPHandler struct {
	service *service.STPService
}

func NewSTPHandler(service *service.STPService) *STPHandler {
	return &STPHandler{service: service}
}

// GetMethods returns 5 random methods for the user to choose from
// @Summary Get random lesson methods
// @Description Returns 5 random methods from the database
// @Tags stp
// @Security ApiKeyAuth
// @Success 200 {array} domain.LessonMethod
// @Router /api/v1/education/stp/methods [get]
func (h *STPHandler) GetMethods(c *fiber.Ctx) error {
	methods, err := h.service.GetMethodsForUser(c.Context())
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_METHODS_FAILED", err.Error())
	}
	return response.Success(c, methods)
}

// GenerateSTP enqueues a new STP generation task
// @Summary Generate a Short Term Plan (ҚМЖ)
// @Description Enqueues a task to generate a lesson plan based on subject, grade, and topic
// @Tags stp
// @Security ApiKeyAuth
// @Param request body domain.STPRequestDTO true "STP Generation Request"
// @Success 202 {object} map[string]string
// @Router /api/v1/education/stp/generate [post]
func (h *STPHandler) GenerateSTP(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	dto := new(domain.STPRequestDTO)
	if err := c.BodyParser(dto); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	id, err := h.service.GenerateSTP(c.Context(), userID, dto)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "GENERATE_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"id":     id,
		"status": "pending",
	})
}

// GetSTP returns the status and content of a generated STP
// @Summary Get STP status and content
// @Description Returns the details of a generated lesson plan
// @Tags stp
// @Security ApiKeyAuth
// @Param id path string true "STP ID"
// @Success 200 {object} domain.ShortTermPlan
// @Router /api/v1/education/stp/{id} [get]
func (h *STPHandler) GetSTP(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid ID format")
	}

	stp, err := h.service.GetSTPStatus(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", fmt.Sprintf("STP not found: %s", id.String()))
	}

	return response.Success(c, stp)
}
