package handler

import (
	"strconv"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	service *service.AuditService
}

func NewAuditHandler(s *service.AuditService) *AuditHandler {
	return &AuditHandler{service: s}
}

// GetLogs godoc
// @Summary Get audit logs
// @Description Get system audit logs (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param limit query int false "Limit" default(50)
// @Param offset query int false "Offset" default(0)
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]domain.AuditLog}
// @Failure 401 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Router /admin/logs [get]
func (h *AuditHandler) GetLogs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	logs, err := h.service.GetLogs(c.Context(), limit, offset)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "LOGS_FETCH_FAILED", err.Error())
	}

	return response.Success(c, logs)
}
