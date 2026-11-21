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

type GroupTaskHandler struct {
	service   service.GroupTaskService
	validator *validator.Validate
}

func NewGroupTaskHandler(service service.GroupTaskService) *GroupTaskHandler {
	return &GroupTaskHandler{
		service:   service,
		validator: validation.NewValidator(),
	}
}

func (h *GroupTaskHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	group := api.Group("/group-tasks", middleware.Protected(cfg), middleware.RBAC(domain.RoleTeacher))

	group.Post("", h.CreateDraft)
	group.Post("/:id/generate", h.Generate)
	group.Get("", h.ListMyTasks)
	group.Get("/:id", h.GetTask)
	group.Delete("/:id", h.DeleteTask)

	// Question editing
	group.Post("/:id/questions", h.AddQuestion)
	group.Patch("/:id/questions/:qid", h.UpdateQuestion)
	group.Delete("/:id/questions/:qid", h.DeleteQuestion)

	// Session management
	group.Post("/sessions", h.StartSession)
	group.Get("/sessions/:id", h.GetSession)
	group.Post("/sessions/:id/groups/:groupId/award", h.AwardPoint)
}

func (h *GroupTaskHandler) CreateDraft(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	var req service.CreateGroupTaskRequest
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

	task, err := h.service.CreateDraft(c.Context(), &req, userID, schoolID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Created(c, task)
}

func (h *GroupTaskHandler) Generate(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}

	if err := h.service.GenerateQuestions(c.Context(), id, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Generation started"})
}

func (h *GroupTaskHandler) ListMyTasks(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "20"))

	tasks, err := h.service.ListMyTasks(c.Context(), userID, page, pageSize)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, tasks)
}

func (h *GroupTaskHandler) GetTask(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}

	task, err := h.service.GetTaskWithQuestions(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	}

	return response.Success(c, task)
}

func (h *GroupTaskHandler) DeleteTask(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}

	if err := h.service.DeleteTask(c.Context(), id, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Deleted"})
}

func (h *GroupTaskHandler) AddQuestion(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}

	var req service.QuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	q, err := h.service.AddQuestion(c.Context(), taskID, userID, &req)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}
	return response.Created(c, q)
}

func (h *GroupTaskHandler) UpdateQuestion(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}
	qid, err := uuid.Parse(c.Params("qid"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid question ID")
	}

	var req service.QuestionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}
	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.UpdateQuestion(c.Context(), taskID, qid, userID, &req); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Updated"})
}

func (h *GroupTaskHandler) DeleteQuestion(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	taskID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
	}
	qid, err := uuid.Parse(c.Params("qid"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid question ID")
	}

	if err := h.service.DeleteQuestion(c.Context(), taskID, qid, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}
	return response.Success(c, fiber.Map{"message": "Deleted"})
}

func (h *GroupTaskHandler) StartSession(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	var req service.StartSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
	}

	session, groups, err := h.service.StartSession(c.Context(), &req, userID)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"session": session,
		"groups":  groups,
	})
}

func (h *GroupTaskHandler) GetSession(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID")
	}

	session, groups, err := h.service.GetSession(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, http.StatusNotFound, "NOT_FOUND", err.Error())
	}

	return response.Success(c, fiber.Map{
		"session": session,
		"groups":  groups,
	})
}

func (h *GroupTaskHandler) AwardPoint(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil { return err }

	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid session ID")
	}

	groupID, err := uuid.Parse(c.Params("groupId"))
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid group ID")
	}

	if err := h.service.AwardPoint(c.Context(), sessionID, groupID, userID); err != nil {
		return response.Error(c, http.StatusInternalServerError, "FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{"message": "Point awarded"})
}
