package handler

import (
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

type TugOfWarHandler struct {
	service   service.TugOfWarService
	validator *validator.Validate
}

func NewTugOfWarHandler(service service.TugOfWarService) *TugOfWarHandler {
	return &TugOfWarHandler{
		service:   service,
		validator: validation.NewValidator(),
	}
}

func (h *TugOfWarHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	group := api.Group("/tug-of-war", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher))

	group.Post("", h.CreateDraft)
	group.Post("/:id/generate", h.Generate)
	group.Get("", h.ListMyMaterials)
	group.Get("/:id", h.GetWithQuestions)
	group.Delete("/:id", h.DeleteMaterial)
}

func (h *TugOfWarHandler) CreateDraft(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	var req service.CreateTugOfWarRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	var schoolID *uuid.UUID
	if sid := c.Locals("school_id"); sid != nil {
		id := sid.(uuid.UUID)
		schoolID = &id
	}

	material, err := h.service.CreateDraft(c.Context(), &req, userID, schoolID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Created(c, material)
}

func (h *TugOfWarHandler) Generate(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid material ID")
	}

	if err := h.service.GenerateQuestions(c.Context(), id, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Generation started"})
}

func (h *TugOfWarHandler) ListMyMaterials(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	result, err := h.service.ListMyMaterials(c.Context(), userID, page, pageSize)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, result)
}

func (h *TugOfWarHandler) GetWithQuestions(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid material ID")
	}

	result, err := h.service.GetWithQuestions(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	}

	return response.Success(c, result)
}

func (h *TugOfWarHandler) DeleteMaterial(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid material ID")
	}

	if err := h.service.DeleteMaterial(c.Context(), id, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Deleted"})
}
