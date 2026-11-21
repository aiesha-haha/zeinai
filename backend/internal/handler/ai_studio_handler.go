package handler

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/middleware"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/pkg/validation"
	"zeinai-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// StudioHandler exposes AI Studio (image, video, feedback templates, lab work,
// school news, certificates) under /api/v1/ai-studio.
type StudioHandler struct {
	svc       *service.StudioService
	validator *validator.Validate
}

func NewStudioHandler(svc *service.StudioService) *StudioHandler {
	return &StudioHandler{svc: svc, validator: validation.NewValidator()}
}

func (h *StudioHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	g := api.Group("/ai-studio", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin))
	g.Get("/status", h.Status)
	g.Post("", h.Create)
	g.Get("", h.List)
	g.Get("/:id", h.Get)
	g.Put("/:id", h.Update)
	g.Post("/:id/retry", h.Retry)
	g.Get("/:id/download", h.Download)
	g.Delete("/:id", h.Delete)
}

func (h *StudioHandler) fail(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrStudioNotConfigured):
		return response.Error(c, http.StatusServiceUnavailable, "AI_NOT_CONFIGURED", err.Error())
	case errors.Is(err, service.ErrStudioNotFound):
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	case strings.Contains(err.Error(), "insufficient tokens"):
		return response.Error(c, http.StatusPaymentRequired, "INSUFFICIENT_TOKENS", err.Error())
	}
	return response.Error(c, http.StatusBadRequest, "STUDIO_ERROR", err.Error())
}

// Status tells the UI which kinds can run on this server (API keys present).
func (h *StudioHandler) Status(c *fiber.Ctx) error {
	kinds := []string{
		domain.StudioKindImage, domain.StudioKindVideo, domain.StudioKindFeedback,
		domain.StudioKindLab, domain.StudioKindNews, domain.StudioKindCertificate,
	}
	out := fiber.Map{}
	for _, k := range kinds {
		missing := h.svc.MissingKeyFor(k)
		out[k] = fiber.Map{"available": missing == "", "missing_env": missing}
	}
	return response.Success(c, out)
}

func (h *StudioHandler) Create(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	var req service.CreateStudioRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	item, err := h.svc.Create(c.Context(), userID, &req)
	if err != nil {
		return h.fail(c, err)
	}
	return response.Created(c, item)
}

func (h *StudioHandler) List(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	items, err := h.svc.List(c.Context(), userID, c.Query("kind"))
	if err != nil {
		return h.fail(c, err)
	}
	return response.Success(c, items)
}

func (h *StudioHandler) Get(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ID")
	}
	item, err := h.svc.Get(c.Context(), id, userID)
	if err != nil {
		return h.fail(c, err)
	}
	return response.Success(c, item)
}

func (h *StudioHandler) Update(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ID")
	}
	var req service.UpdateStudioRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	item, err := h.svc.Update(c.Context(), id, userID, &req)
	if err != nil {
		return h.fail(c, err)
	}
	return response.Success(c, item)
}

func (h *StudioHandler) Retry(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ID")
	}
	item, err := h.svc.Retry(c.Context(), id, userID)
	if err != nil {
		return h.fail(c, err)
	}
	return response.Success(c, item)
}

// Download sends the generated file as an attachment (needs the bearer token,
// so the frontend fetches it as a blob).
func (h *StudioHandler) Download(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ID")
	}
	item, err := h.svc.Get(c.Context(), id, userID)
	if err != nil {
		return h.fail(c, err)
	}
	if item.FileURL == nil || *item.FileURL == "" {
		return response.Error(c, http.StatusNotFound, "NO_FILE", "File is not ready")
	}
	name := filepath.Base(*item.FileURL) // e.g. <uuid>.png — never trust a path from the DB
	path := filepath.Join(".", "uploads", "studio", name)
	return c.Download(path, "zeinai-"+item.Kind+filepath.Ext(name))
}

func (h *StudioHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ID")
	}
	if err := h.svc.Delete(c.Context(), id, userID); err != nil {
		return h.fail(c, err)
	}
	return response.Success(c, fiber.Map{"deleted": true})
}
