package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"
)

type AIHandler struct {
	service *service.AIService
}

func NewAIHandler(s *service.AIService) *AIHandler {
	return &AIHandler{service: s}
}

type GenerateRequest struct {
	Prompt string `json:"prompt"`
}

type SuggestInstructionsRequest struct {
	Subject  string `json:"subject"`
	Grade    int    `json:"grade"`
	Topic    string `json:"topic"`
	Language string `json:"language"`
}

// GenerateText godoc
// @Summary Generate text using AI
// @Description Send a prompt to Gemini AI
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.GenerateRequest true "Prompt"
// @Success 201 {object} response.APIResponse{data=domain.AIRequest}
// @Failure 400 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /ai/generate [post]
func (h *AIHandler) GenerateText(c *fiber.Ctx) error {
	var req GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if req.Prompt == "" {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_PROMPT", "Prompt is required")
	}

	// Get UserID from middleware
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	task, err := h.service.EnqueueTextGeneration(c.Context(), userID, req.Prompt)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "QUEUE_FAILED", err.Error())
	}

	return response.Created(c, task)
}

// SuggestArticleInstructions godoc
// @Summary Suggest article instructions
// @Description Generate AI-suggested instructions for an article
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.SuggestInstructionsRequest true "Article Info"
// @Success 201 {object} response.APIResponse{data=domain.AIRequest}
// @Router /ai/suggest-article-instructions [post]
func (h *AIHandler) SuggestArticleInstructions(c *fiber.Ctx) error {
	var req SuggestInstructionsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	task, err := h.service.SuggestArticleInstructions(c.Context(), userID, req.Subject, req.Grade, req.Topic, req.Language)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "QUEUE_FAILED", err.Error())
	}

	return response.Created(c, task)
}

// SuggestPresentationInstructions godoc
// @Summary Suggest presentation instructions
// @Description Generate AI-suggested instructions for a presentation
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body handler.SuggestInstructionsRequest true "Presentation Info"
// @Success 201 {object} response.APIResponse{data=domain.AIRequest}
// @Router /ai/suggest-presentation-instructions [post]
func (h *AIHandler) SuggestPresentationInstructions(c *fiber.Ctx) error {
	var req SuggestInstructionsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	task, err := h.service.SuggestPresentationInstructions(c.Context(), userID, req.Subject, req.Grade, req.Topic, req.Language)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "QUEUE_FAILED", err.Error())
	}

	return response.Created(c, task)
}

// GetStatus godoc
// @Summary Get AI task status
// @Description Get status and result of an AI task
// @Tags AI
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID"
// @Success 200 {object} response.APIResponse{data=domain.AIRequest}
// @Failure 404 {object} response.APIResponse
// @Router /ai/status/{id} [get]
func (h *AIHandler) GetStatus(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid Task ID")
	}

	task, err := h.service.GetRequest(c.Context(), id)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DB_ERROR", err.Error())
	}
	if task == nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", "Task not found")
	}

	return response.Success(c, task)
}
