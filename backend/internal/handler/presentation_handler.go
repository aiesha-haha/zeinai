package handler

import (
	"log"
	"net/http"
	"strconv"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/middleware"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// PresentationHandler handles all /presentations endpoints.
type PresentationHandler struct {
	service   service.PresentationService
	validator *validator.Validate
}

func NewPresentationHandler(svc service.PresentationService) *PresentationHandler {
	return &PresentationHandler{
		service:   svc,
		validator: validation.NewValidator(),
	}
}

// RegisterRoutes mounts the presentation routes onto the API router.
func (h *PresentationHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	g := api.Group("/presentations", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher))

	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Get("/:id", h.GetByID)
	g.Delete("/:id", h.Delete)
}

// ──────────────────────────────────────────────────────────

// Create godoc
// @Summary      Create and generate a presentation
// @Description  Creates a new AI-generated presentation (slides) using Kimi K2. Generation is async – poll GET /presentations/:id to check progress.
// @Tags         Presentations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body service.CreatePresentationRequest true "Presentation generation request"
// @Success      202 {object} response.APIResponse{data=domain.Presentation}   "Generation started"
// @Failure      400 {object} response.APIResponse{error=response.APIError}    "Validation error"
// @Failure      401 {object} response.APIResponse{error=response.APIError}    "Unauthorized"
// @Failure      403 {object} response.APIResponse{error=response.APIError}    "Insufficient tokens or wrong role"
// @Failure      500 {object} response.APIResponse{error=response.APIError}    "Server error"
// @Router       /api/v1/presentations [post]
func (h *PresentationHandler) Create(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var schoolID *int64
	if raw := c.Locals("school_id"); raw != nil {
		switch v := raw.(type) {
		case int64:
			schoolID = &v
		case *int64:
			schoolID = v
		}
	}

	var req service.CreatePresentationRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	p, err := h.service.Create(c.Context(), &req, userID, schoolID)
	if err != nil {
		log.Printf("[PresentationHandler] Create error: %v", err)
		if containsMsg(err, "insufficient tokens") {
			return response.Error(c, http.StatusForbidden, "INSUFFICIENT_TOKENS", "Not enough tokens to generate a presentation")
		}
		return response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to start presentation generation", err.Error())
	}

	return c.Status(http.StatusAccepted).JSON(response.APIResponse{
		Success: true,
		Data:    p,
	})
}

// GetByID godoc
// @Summary      Get presentation by ID (with slides)
// @Description  Returns a presentation and all its slides. Poll this endpoint to check generation status.
// @Tags         Presentations
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Presentation ID (UUID)"
// @Success      200 {object} response.APIResponse{data=domain.PresentationWithSlides}
// @Failure      400 {object} response.APIResponse{error=response.APIError}
// @Failure      403 {object} response.APIResponse{error=response.APIError}
// @Failure      404 {object} response.APIResponse{error=response.APIError}
// @Router       /api/v1/presentations/{id} [get]
func (h *PresentationHandler) GetByID(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid presentation ID format")
	}

	result, err := h.service.GetByID(c.Context(), id, userID)
	if err != nil {
		log.Printf("[PresentationHandler] GetByID error: %v", err)
		if containsMsg(err, "not found") {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Presentation not found")
		}
		if containsMsg(err, "unauthorized") {
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "You do not own this presentation")
		}
		return response.Error(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch presentation", err.Error())
	}

	return response.Success(c, result)
}

// List godoc
// @Summary      List my presentations
// @Description  Returns a paginated list of the current user's presentations.
// @Tags         Presentations
// @Produce      json
// @Security     BearerAuth
// @Param        page      query int false "Page (default: 1)"
// @Param        page_size query int false "Page size (default: 20, max: 100)"
// @Success      200 {object} response.APIResponse{data=service.PaginatedPresentations}
// @Failure      401 {object} response.APIResponse{error=response.APIError}
// @Router       /api/v1/presentations [get]
func (h *PresentationHandler) List(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	result, err := h.service.List(c.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("[PresentationHandler] List error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list presentations", err.Error())
	}

	return response.Success(c, result)
}

// Delete godoc
// @Summary      Delete a presentation
// @Description  Soft-deletes a presentation and all associated slides.
// @Tags         Presentations
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Presentation ID (UUID)"
// @Success      200 {object} response.APIResponse{data=object{message=string}}
// @Failure      400 {object} response.APIResponse{error=response.APIError}
// @Failure      403 {object} response.APIResponse{error=response.APIError}
// @Router       /api/v1/presentations/{id} [delete]
func (h *PresentationHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid presentation ID format")
	}

	if err := h.service.Delete(c.Context(), id, userID); err != nil {
		log.Printf("[PresentationHandler] Delete error: %v", err)
		if containsMsg(err, "unauthorized") {
			return response.Error(c, http.StatusForbidden, "FORBIDDEN", "You do not own this presentation")
		}
		if containsMsg(err, "not found") {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Presentation not found")
		}
		return response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete presentation", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Presentation deleted successfully"})
}

// containsMsg is a small helper to check if an error message contains a substring.
func containsMsg(err error, sub string) bool {
	if err == nil {
		return false
	}
	return len(err.Error()) > 0 && len(sub) > 0 && stringContains(err.Error(), sub)
}

func stringContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
