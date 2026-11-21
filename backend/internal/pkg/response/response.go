package response

import (
	"github.com/gofiber/fiber/v2"
)

// APIResponse is the standard response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"` // For pagination etc
}

// APIError defines the structure for error details
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Success sends a success response with 200 OK or specific status
func Success(c *fiber.Ctx, data interface{}, meta ...interface{}) error {
	resp := APIResponse{
		Success: true,
		Data:    data,
	}
	if len(meta) > 0 {
		resp.Meta = meta[0]
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

// Created sends a success response with 201 Created
func Created(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *fiber.Ctx, status int, code string, message string, details ...interface{}) error {
	resp := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	}
	if len(details) > 0 {
		resp.Error.Details = details[0]
	}
	return c.Status(status).JSON(resp)
}
