package handler

import (
	"context"
	"net/http/httptest"
	"testing"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// MockKMJService is a simple mock
type MockKMJService struct {
	service.KMJService
}

func (m *MockKMJService) ListTeachingMethods(ctx context.Context) ([]domain.TeachingMethod, error) {
	return []domain.TeachingMethod{
		{ID: uuid.New(), NameKaz: "Тест әдіс"},
	}, nil
}

func TestKMJHandler_ListTeachingMethods(t *testing.T) {
	app := fiber.New()
	mockService := &MockKMJService{}
	h := NewKMJHandler(mockService)

	app.Get("/teaching-methods", h.ListTeachingMethods)

	req := httptest.NewRequest("GET", "/teaching-methods", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}
