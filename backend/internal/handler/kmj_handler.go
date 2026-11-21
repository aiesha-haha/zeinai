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

type KMJHandler struct {
	kmjService service.KMJService
	validator  *validator.Validate
}

func NewKMJHandler(kmjService service.KMJService) *KMJHandler {
	return &KMJHandler{
		kmjService: kmjService,
		validator:  validation.NewValidator(),
	}
}

// RegisterRoutes registers all KMJ routes
func (h *KMJHandler) RegisterRoutes(api fiber.Router, cfg *config.Config) {
	kmj := api.Group("/kmj", middleware.Protected(cfg))

	// Teacher only endpoints
	kmj.Post("/", middleware.RBAC(domain.RoleTeacher), h.CreateDraftPlan)
	kmj.Post("/:id/generate", middleware.RBAC(domain.RoleTeacher), h.GeneratePlan)
	kmj.Get("/", middleware.RBAC(domain.RoleTeacher), h.ListMyPlans)
	kmj.Get("/:id", middleware.RBAC(domain.RoleTeacher), h.GetPlan)
	kmj.Delete("/:id", middleware.RBAC(domain.RoleTeacher), h.DeletePlan)

	// Cell editing
	kmj.Put("/cells/:cellId", middleware.RBAC(domain.RoleTeacher), h.UpdateCell)

	// Additional public data endpoints (within KMJ context)
	api.Get("/teaching-methods", h.ListTeachingMethods)
	api.Get("/national-values/month/:month", h.GetNationalValueByMonth)
}

// Helper function to extract user ID from context
func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return uuid.UUID{}, fiber.NewError(fiber.StatusUnauthorized, "user_id not found in context")
	}
	return userID, nil
}

// CreateDraftPlan godoc
// @Summary Create a draft KMJ lesson plan
// @Description Creates a new draft lesson plan (Қысқа мерзімді жоспар) with empty structure cells. All cells are created and can be edited individually. This endpoint does NOT consume tokens - only generation does.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body service.CreateKMJRequest true "Lesson plan details"
// @Success 201 {object} response.APIResponse{data=domain.KMJPlanWithCells} "Draft plan created successfully with empty cells"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid request body or validation failed"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - requires teacher role"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/kmj [post]
func (h *KMJHandler) CreateDraftPlan(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var schoolID *uuid.UUID
	if schoolIDLocal := c.Locals("school_id"); schoolIDLocal != nil {
		if sid, ok := schoolIDLocal.(uuid.UUID); ok {
			schoolID = &sid
		}
	}

	var req service.CreateKMJRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[KMJ] Body parse error: %v", err)
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
	}

	if err := h.validator.Struct(&req); err != nil {
		log.Printf("[KMJ] Validation error: %v", err)
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	planWithCells, err := h.kmjService.CreateDraftPlan(c.Context(), &req, userID, schoolID)
	if err != nil {
		log.Printf("[KMJ] Create draft error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create draft plan", err.Error())
	}

	return response.Created(c, planWithCells)
}

// GeneratePlan godoc
// @Summary Generate lesson plan content with AI
// @Description Starts asynchronous AI generation for a draft lesson plan. **IMPORTANT: Consumes 5 tokens from user balance.** The generation happens in background via queue. Check plan status to see when it's completed.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Plan ID (UUID)" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} response.APIResponse{data=object{message=string,plan_id=string}} "Generation started successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid plan ID format"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - not plan owner OR insufficient tokens (need 5)"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Plan not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error or queue publishing failed"
// @Router /api/v1/kmj/{id}/generate [post]
//
//	-H "Authorization: Bearer YOUR_TOKEN"
//	'/api/v1/kmj/$planId/generate',
//	);
func (h *KMJHandler) GeneratePlan(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	planIDStr := c.Params("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", "Invalid plan ID format")
	}

	err = h.kmjService.GenerateKMJ(c.Context(), planID, userID)
	if err != nil {
		log.Printf("[KMJ] Generate error: %v", err)
		if err.Error() == "unauthorized: not plan owner" {
			return response.Error(c, http.StatusForbidden, "NOT_PLAN_OWNER", "You are not the owner of this plan")
		}
		return response.Error(c, http.StatusInternalServerError, "GENERATION_FAILED", "Failed to start generation", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Generation started successfully",
		"plan_id": planID,
	})
}

// GetPlan godoc
// @Summary Get lesson plan by ID
// @Description Retrieves a complete lesson plan with all cells. Use this to check generation status and view/edit cells.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Plan ID (UUID)" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} response.APIResponse{data=domain.KMJPlanWithCells} "Plan retrieved successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid plan ID format"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - not plan owner"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Plan not found"
// @Router /api/v1/kmj/{id} [get]
//
//	'/api/v1/kmj/$planId',
//	);
func (h *KMJHandler) GetPlan(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	planIDStr := c.Params("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", "Invalid plan ID format")
	}

	planWithCells, err := h.kmjService.GetPlanByID(c.Context(), planID, userID)
	if err != nil {
		log.Printf("[KMJ] Get plan error: %v", err)
		if err.Error() == "unauthorized: not plan owner" {
			return response.Error(c, http.StatusForbidden, "NOT_PLAN_OWNER", "You are not the owner of this plan")
		}
		return response.Error(c, http.StatusNotFound, "PLAN_NOT_FOUND", "Plan not found")
	}

	return response.Success(c, planWithCells)
}

// ListMyPlans godoc
// @Summary List user's lesson plans
// @Description Get paginated list of current user's lesson plans. Supports filtering by status.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" minimum(1) default(1)
// @Param page_size query int false "Items per page (default: 50, max: 100)" minimum(1) maximum(100) default(50)
// @Success 200 {object} response.APIResponse{data=service.PaginatedKMJPlans} "Plans list retrieved successfully"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - requires teacher role"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/kmj [get]
//
//	'/api/v1/kmj',
//	queryParameters: {'page': 1, 'page_size': 50},
//	);
func (h *KMJHandler) ListMyPlans(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))

	result, err := h.kmjService.ListMyPlans(c.Context(), userID, page, pageSize)
	if err != nil {
		log.Printf("[KMJ] List plans error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "LIST_FAILED", "Failed to list plans", err.Error())
	}

	return response.Success(c, result)
}

// DeletePlan godoc
// @Summary Delete lesson plan
// @Description Soft deletes a lesson plan. The plan and all its cells are marked as deleted but remain in database.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Plan ID (UUID)" example(123e4567-e89b-12d3-a456-426614174000)
// @Success 200 {object} response.APIResponse{data=object{message=string}} "Plan deleted successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid plan ID format"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - not plan owner"
// @Failure 404 {object} response.APIResponse{error=response.APIError} "Plan not found"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/kmj/{id} [delete]
func (h *KMJHandler) DeletePlan(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	planIDStr := c.Params("id")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_PLAN_ID", "Invalid plan ID format")
	}

	err = h.kmjService.DeletePlan(c.Context(), planID, userID)
	if err != nil {
		log.Printf("[KMJ] Delete error: %v", err)
		if err.Error() == "unauthorized: not plan owner" {
			return response.Error(c, http.StatusForbidden, "NOT_PLAN_OWNER", "You are not the owner of this plan")
		}
		return response.Error(c, http.StatusInternalServerError, "DELETE_FAILED", "Failed to delete plan", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Plan deleted successfully",
	})
}

// UpdateCell godoc
// @Summary Update lesson plan cell
// @Description Updates the content of a single cell in the lesson plan. Cells can be edited individually. When edited, is_edited flag is set to true.
// @Tags KMJ (Lesson Plans)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param cellId path string true "Cell ID (UUID)" example(123e4567-e89b-12d3-a456-426614174000)
// @Param body body object{cell_value=string} true "New cell content" SchemaExample({"cell_value": "Updated content for this cell"})
// @Success 200 {object} response.APIResponse{data=object{message=string}} "Cell updated successfully"
// @Failure 400 {object} response.APIResponse{error=response.APIError} "Invalid cell ID or request body"
// @Failure 401 {object} response.APIResponse{error=response.APIError} "Unauthorized - missing or invalid token"
// @Failure 403 {object} response.APIResponse{error=response.APIError} "Forbidden - not plan owner"
// @Failure 500 {object} response.APIResponse{error=response.APIError} "Internal server error"
// @Router /api/v1/kmj/cells/{cellId} [put]
//
//	'/api/v1/kmj/cells/$cellId',
//	data: {'cell_value': 'New content'},
//	);
func (h *KMJHandler) UpdateCell(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	cellIDStr := c.Params("cellId")
	cellID, err := uuid.Parse(cellIDStr)
	if err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_CELL_ID", "Invalid cell ID format")
	}

	var req struct {
		CellValue string `json:"cell_value" validate:"required" example:"Обновленное содержимое ячейки"`
	}

	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
	}

	if err := h.validator.Struct(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", err.Error())
	}

	err = h.kmjService.UpdateCell(c.Context(), cellID, req.CellValue, userID)
	if err != nil {
		log.Printf("[KMJ] Update cell error: %v", err)
		return response.Error(c, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update cell", err.Error())
	}

	return response.Success(c, fiber.Map{
		"message": "Cell updated successfully",
	})
}

// ListTeachingMethods godoc
// @Summary List all teaching methods
// @Description Returns a list of all pedagogical methods (50 әдіс) available for lesson planning.
// @Tags KMJ (Lesson Plans)
// @Produce json
// @Success 200 {object} response.APIResponse{data=[]domain.TeachingMethod} "Methods list retrieved successfully"
// @Router /api/v1/teaching-methods [get]
func (h *KMJHandler) ListTeachingMethods(c *fiber.Ctx) error {
	methods, err := h.kmjService.ListTeachingMethods(c.Context())
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch teaching methods")
	}
	return response.Success(c, methods)
}

// GetNationalValueByMonth godoc
// @Summary Get national values by month
// @Description Returns national values (Адал азамат құндылықтары) for a specific month (1-12).
// @Tags KMJ (Lesson Plans)
// @Produce json
// @Param month path int true "Month number (1-12)" example(9)
// @Success 200 {object} response.APIResponse{data=[]domain.NationalValue} "National values retrieved successfully"
// @Router /api/v1/national-values/month/{month} [get]
func (h *KMJHandler) GetNationalValueByMonth(c *fiber.Ctx) error {
	monthStr := c.Params("month")
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		return response.Error(c, http.StatusBadRequest, "INVALID_MONTH", "Invalid month number")
	}

	values, err := h.kmjService.GetNationalValueByMonth(c.Context(), month)
	if err != nil {
		return response.Error(c, http.StatusInternalServerError, "FETCH_FAILED", "Failed to fetch national values")
	}
	return response.Success(c, values)
}
