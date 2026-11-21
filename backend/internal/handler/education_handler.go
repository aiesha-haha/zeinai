package handler

import (
	"github.com/gofiber/fiber/v2"
	"zeinai-backend/internal/pkg/response"
)

type EducationHandler struct{}

func NewEducationHandler() *EducationHandler {
	return &EducationHandler{}
}

func (h *EducationHandler) Placeholder(c *fiber.Ctx) error {
	return response.Success(c, fiber.Map{
		"message": "This feature is under development",
		"status":  "coming_soon",
	})
}
