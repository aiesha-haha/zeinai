package handler

import (
	"strconv"

	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ClassHandler struct {
	service   *service.ClassService
	validator *validator.Validate
}

func NewClassHandler(s *service.ClassService) *ClassHandler {
	return &ClassHandler{
		service:   s,
		validator: validation.NewValidator(),
	}
}

// CreateClass godoc
// @Summary Create a new class
// @Description Create a class for the authenticated teacher
// @Tags Classes
// @Accept json
// @Produce json
// @Param request body domain.CreateClassRequest true "Create Class Request"
// @Success 201 {object} response.APIResponse{data=domain.Class}
// @Failure 400 {object} response.APIResponse
// @Router /classes [post]
func (h *ClassHandler) CreateClass(c *fiber.Ctx) error {
	var req domain.CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	class, err := h.service.CreateClass(c.Context(), req, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_FAILED", err.Error())
	}

	return response.Created(c, class)
}

// ListClasses godoc
// @Summary List classes
// @Description Get all classes for the authenticated teacher
// @Tags Classes
// @Produce json
// @Success 200 {object} response.APIResponse{data=[]domain.Class}
// @Router /classes [get]
func (h *ClassHandler) ListClasses(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}
	classes, err := h.service.GetClasses(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}
	return response.Success(c, classes)
}

// GetClass godoc
// @Summary Get class details
// @Description Get details of a specific class
// @Tags Classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} response.APIResponse{data=domain.Class}
// @Failure 404 {object} response.APIResponse
// @Router /classes/{id} [get]
func (h *ClassHandler) GetClass(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	class, err := h.service.GetClass(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", err.Error())
	}
	return response.Success(c, class)
}

// UpdateClass godoc
// @Summary Update class
// @Description Update class name
// @Tags Classes
// @Accept json
// @Produce json
// @Param id path int true "Class ID"
// @Param request body domain.CreateClassRequest true "Update Class Request"
// @Success 200 {object} response.APIResponse{data=domain.Class}
// @Router /classes/{id} [put]
func (h *ClassHandler) UpdateClass(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.CreateClassRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	class, err := h.service.UpdateClass(c.Context(), id, req, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPDATE_FAILED", err.Error())
	}

	return response.Success(c, class)
}

// DeleteClass godoc
// @Summary Delete class
// @Description Soft delete a class
// @Tags Classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} response.APIResponse
// @Router /classes/{id} [delete]
func (h *ClassHandler) DeleteClass(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	if err := h.service.DeleteClass(c.Context(), id, userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DELETE_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// AddStudent godoc
// @Summary Add student to class
// @Description Create a new student (or link existing) and add to class. Requires unique IIN.
// @Tags Classes
// @Accept json
// @Produce json
// @Param id path int true "Class ID"
// @Param request body domain.AddStudentRequest true "Add Student Request"
// @Success 200 {object} response.APIResponse{data=domain.User}
// @Failure 400 {object} response.APIResponse
// @Router /classes/{id}/students [post]
func (h *ClassHandler) AddStudent(c *fiber.Ctx) error {
	classID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	var req domain.AddStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	// Ensure URL param matches body if body has class_id, or set it
	req.ClassID = classID

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	student, password, err := h.service.AddStudent(c.Context(), req, userID)
	if err != nil {
		if err.Error() == "student limit reached. please upgrade your plan" {
			return response.Error(c, fiber.StatusForbidden, "LIMIT_REACHED", err.Error())
		}
		return response.Error(c, fiber.StatusInternalServerError, "ADD_STUDENT_FAILED", err.Error())
	}

	return response.Success(c, fiber.Map{
		"student":  student,
		"password": password, // Non-empty only for new users
	})
}

// ListStudents godoc
// @Summary List students in class
// @Description Get all students in a specific class
// @Tags Classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} response.APIResponse{data=[]domain.User}
// @Router /classes/{id}/students [get]
func (h *ClassHandler) ListStudents(c *fiber.Ctx) error {
	id, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	students, err := h.service.GetClassStudents(c.Context(), id, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}
	return response.Success(c, students)
}

// RemoveStudent godoc
// @Summary Remove student from class
// @Description Remove a student from a class
// @Tags Classes
// @Produce json
// @Param id path int true "Class ID"
// @Param student_id path string true "Student ID (UUID)"
// @Success 200 {object} response.APIResponse
// @Router /classes/{id}/students/{student_id} [delete]
func (h *ClassHandler) RemoveStudent(c *fiber.Ctx) error {
	classID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	studentID, err := uuid.Parse(c.Params("student_id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UUID", "Invalid student ID format")
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	if err := h.service.RemoveStudent(c.Context(), classID, studentID, userID); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "REMOVE_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// CreateInvite godoc
// @Summary Create invite link for class
// @Description Generate a unique code for students to join the class
// @Tags Classes
// @Accept json
// @Produce json
// @Param id path int true "Class ID"
// @Param request body domain.CreateInviteRequest true "Create Invite Request"
// @Success 201 {object} response.APIResponse{data=domain.ClassInvite}
// @Router /classes/{id}/invites [post]
func (h *ClassHandler) CreateInvite(c *fiber.Ctx) error {
	classID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	var req domain.CreateInviteRequest

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	invite, err := h.service.CreateInvite(c.Context(), classID, userID, req)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "CREATE_INVITE_FAILED", err.Error())
	}

	return response.Created(c, invite)
}

// GetActiveInvite godoc
// @Summary Get active invite
// @Description Get the currently active invite for a class (Teacher only)
// @Tags Classes
// @Produce json
// @Param id path int true "Class ID"
// @Success 200 {object} response.APIResponse{data=domain.ClassInvite}
// @Router /classes/{id}/invites/active [get]
func (h *ClassHandler) GetActiveInvite(c *fiber.Ctx) error {
	classID, _ := strconv.ParseInt(c.Params("id"), 10, 64)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	invite, err := h.service.GetActiveInvite(c.Context(), classID, userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "FETCH_FAILED", err.Error())
	}

	// It can be nil if no active invite exists
	if invite == nil {
		return response.Success(c, nil)
	}

	return response.Success(c, invite)
}

// GetInviteInfo godoc
// @Summary Get invite info
// @Description Get class name and teacher from invite code (public)
// @Tags Invites
// @Produce json
// @Param code path string true "Invite Code"
// @Success 200 {object} response.APIResponse{data=domain.InviteInfoResponse}
// @Failure 404 {object} response.APIResponse
// @Router /invites/{code} [get]
func (h *ClassHandler) GetInviteInfo(c *fiber.Ctx) error {
	code := c.Params("code")
	info, err := h.service.GetInviteInfo(c.Context(), code)
	if err != nil {
		if err.Error() == "invite expired" || err.Error() == "invite limit reached" {
			return response.Error(c, fiber.StatusGone, "INVITE_INVALID", err.Error())
		}
		return response.Error(c, fiber.StatusNotFound, "NOT_FOUND", err.Error())
	}
	return response.Success(c, info)
}

// AcceptInvite godoc
// @Summary Accept class invite
// @Description Join a class using an invite code. Requires authentication.
// @Tags Invites
// @Produce json
// @Param code path string true "Invite Code"
// @Success 200 {object} response.APIResponse
// @Router /invites/{code}/accept [post]
func (h *ClassHandler) AcceptInvite(c *fiber.Ctx) error {
	code := c.Params("code")
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	if err := h.service.AcceptInvite(c.Context(), code, userID); err != nil {
		if err.Error() == "invite limit reached" || err.Error() == "teacher's student limit reached" {
			return response.Error(c, fiber.StatusForbidden, "LIMIT_REACHED", err.Error())
		}
		if err.Error() == "invite expired" {
			return response.Error(c, fiber.StatusGone, "INVITE_EXPIRED", err.Error())
		}
		return response.Error(c, fiber.StatusBadRequest, "JOIN_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// GetStudentDashboard godoc
// @Summary Get student dashboard
// @Description Get student profile and classes
// @Tags Student
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.StudentDashboardResponse}
// @Failure 401 {object} response.APIResponse
// @Router /student/dashboard [get]
func (h *ClassHandler) GetStudentDashboard(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	resp, err := h.service.GetStudentDashboard(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DASHBOARD_ERROR", err.Error())
	}

	return response.Success(c, resp)
}

// NotifyStudent godoc
// @Summary Notify a specific student
// @Description Send a system notification to a specific student (Teacher only)
// @Tags Teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.TeacherNotifyStudentRequest true "Notification Data"
// @Success 200 {object} response.APIResponse
// @Router /teacher/notify-student [post]
func (h *ClassHandler) NotifyStudent(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}
	var req domain.TeacherNotifyStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.NotifyStudent(c.Context(), userID, req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "NOTIFY_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// NotifyClass godoc
// @Summary Notify all students in a class
// @Description Send a system notification to all students in a class (Teacher only)
// @Tags Teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.TeacherNotifyClassRequest true "Notification Data"
// @Success 200 {object} response.APIResponse
// @Router /teacher/notify-class [post]
func (h *ClassHandler) NotifyClass(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}
	var req domain.TeacherNotifyClassRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.NotifyClass(c.Context(), userID, req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "NOTIFY_FAILED", err.Error())
	}

	return response.Success(c, nil)
}

// NotifyAllStudents godoc
// @Summary Notify all students of a teacher
// @Description Send a system notification to all students across all classes (Teacher only)
// @Tags Teachers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.TeacherNotifyAllStudentsRequest true "Notification Data"
// @Success 200 {object} response.APIResponse
// @Router /teacher/notify-all-students [post]
func (h *ClassHandler) NotifyAllStudents(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}
	var req domain.TeacherNotifyAllStudentsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.NotifyAllStudents(c.Context(), userID, req); err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "NOTIFY_FAILED", err.Error())
	}

	return response.Success(c, nil)
}
