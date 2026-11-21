package handler

import (
	"strconv"

	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/middleware"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"zeinai-backend/internal/pkg/validation"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TestHandler struct {
	svc       service.TestService
	validator *validator.Validate
}

func NewTestHandler(svc service.TestService) *TestHandler {
	return &TestHandler{svc: svc, validator: validation.NewValidator()}
}

func schoolIDFromCtx(c *fiber.Ctx) *int64 {
	switch v := c.Locals("school_id").(type) {
	case int64:
		return &v
	case *int64:
		return v
	}
	return nil
}

// CreateManualTest creates a ready test from supplied questions (no AI, no tokens).
func (h *TestHandler) CreateManualTest(c *fiber.Ctx) error {
	var req service.CreateManualTestRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}
	t, err := h.svc.CreateManualTest(c.Context(), &req, userID, schoolIDFromCtx(c))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "CREATE_FAILED", err.Error())
	}
	return response.Created(c, t)
}

func (h *TestHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	tests := api.Group("/tests", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher, domain.RoleAdmin))
	tests.Post("", h.CreateTest)
	tests.Post("/manual", h.CreateManualTest)
	tests.Get("", h.ListMyTests)
	tests.Get("/:id", h.GetTest)
	tests.Delete("/:id", h.DeleteTest)
	tests.Post("/:id/assign", h.AssignToClass)
	tests.Get("/assignments/all", h.ListTeacherAssignments)
	tests.Get("/assignments/:id/results", h.GetAssignmentResults)
	tests.Delete("/assignments/:id", h.DeleteAssignment)
}

func (h *TestHandler) RegisterStudentRoutes(student fiber.Router) {
	student.Get("/tests", h.ListStudentTests)
	student.Get("/tests/:id", h.GetStudentTest)
	student.Post("/tests/:assignmentID/submit", h.SubmitTest)
}

func (h *TestHandler) CreateTest(c *fiber.Ctx) error {
	var req service.CreateTestRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	schoolIDVal := c.Locals("school_id")
	var schoolID *int64
	if schoolIDVal != nil {
		switch v := schoolIDVal.(type) {
		case int64:
			schoolID = &v
		case *int64:
			schoolID = v
		}
	}

	t, err := h.svc.CreateTest(c.Context(), &req, userID, schoolID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Created(c, t)
}

func (h *TestHandler) GetTest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid test ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	res, err := h.svc.GetWithQuestions(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, res)
}

func (h *TestHandler) ListMyTests(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))

	res, err := h.svc.ListMyTests(c.Context(), userID, page, pageSize)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, res)
}

func (h *TestHandler) DeleteTest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid test ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	err = h.svc.DeleteTest(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, nil)
}

func (h *TestHandler) AssignToClass(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid test ID")
	}

	var req struct {
		ClassID int64 `json:"class_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	err = h.svc.AssignToClass(c.Context(), id, userID, req.ClassID)
	if err != nil {
		if err.Error() == "unauthorized or test not found" || err.Error() == "unauthorized or class not found" {
			return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
		}
		if err.Error() == "test already assigned to this class" {
			return response.Error(c, fiber.StatusBadRequest, "ALREADY_ASSIGNED", err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, nil)
}

func (h *TestHandler) ListStudentTests(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	tests, err := h.svc.ListStudentTests(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, tests)
}

func (h *TestHandler) GetStudentTest(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid test ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	res, err := h.svc.GetTestForStudent(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, res)
}

func (h *TestHandler) ListTeacherAssignments(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	assignments, err := h.svc.ListTeacherAssignments(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, assignments)
}

func (h *TestHandler) GetAssignmentResults(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	results, err := h.svc.GetAssignmentResults(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, results)
}

func (h *TestHandler) DeleteAssignment(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	err = h.svc.DeleteAssignment(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, nil)
}
func (h *TestHandler) SubmitTest(c *fiber.Ctx) error {
	assignmentIDParam := c.Params("assignmentID")
	assignmentID, err := uuid.Parse(assignmentIDParam)
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_ID", "Invalid assignment ID")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
	}

	var req domain.SubmitTestRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
	}

	res, err := h.svc.SubmitTest(c.Context(), assignmentID, userID, &req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, res)
}
