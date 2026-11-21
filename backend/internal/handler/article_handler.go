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

type ArticleHandler struct {
	service   service.ArticleService
	validator *validator.Validate
}

func NewArticleHandler(svc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{
		service:   svc,
		validator: validation.NewValidator(),
	}
}

func (h *ArticleHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	g := api.Group("/articles", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher))

	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Get("/:id", h.GetByID)
	g.Delete("/:id", h.Delete)
}

func (h *ArticleHandler) Create(c *fiber.Ctx) error {
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

	var req domain.CreateArticleRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	a, err := h.service.Create(c.Context(), &req, userID, schoolID)
	if err != nil {
		log.Printf("[ArticleHandler] Create error: %v", err)
		if containsMsg(err, "insufficient tokens") {
			return response.Error(c, http.StatusForbidden, "INSUFFICIENT_TOKENS", "Not enough tokens to generate an article")
		}
		return response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to start article generation", err.Error())
	}

	return c.Status(http.StatusAccepted).JSON(response.APIResponse{
		Success: true,
		Data:    a,
	})
}

func (h *ArticleHandler) GetByID(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid article ID format")
	}

	result, err := h.service.GetByID(c.Context(), id, userID)
	if err != nil {
		log.Printf("[ArticleHandler] GetByID error: %v", err)
		if containsMsg(err, "not found") {
			return response.Error(c, http.StatusNotFound, "NOT_FOUND", "Article not found")
		}
		return response.Error(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch article", err.Error())
	}

	return response.Success(c, result)
}

func (h *ArticleHandler) List(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	result, err := h.service.List(c.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("[ArticleHandler] List error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list articles", err.Error())
	}

	return response.Success(c, result)
}

func (h *ArticleHandler) Delete(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid article ID format")
	}

	if err := h.service.Delete(c.Context(), id, userID); err != nil {
		log.Printf("[ArticleHandler] Delete error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete article", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Article deleted successfully"})
}
