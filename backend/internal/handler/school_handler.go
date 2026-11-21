package handler

import (
	"github.com/gofiber/fiber/v2"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
)

type SchoolHandler struct {
	repo domain.SchoolRepository
}

func NewSchoolHandler(repo domain.SchoolRepository) *SchoolHandler {
	return &SchoolHandler{repo: repo}
}

// List godoc
// @Summary List schools
// @Description List schools with filtering and pagination
// @Tags Schools
// @Accept json
// @Produce json
// @Param region_id query int false "Region ID"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param search query string false "Search query"
// @Success 200 {object} response.APIResponse{data=[]domain.School}
// @Failure 500 {object} response.APIResponse
// @Router /schools [get]
func (h *SchoolHandler) List(c *fiber.Ctx) error {
	regionID := c.QueryInt("region_id", 0)
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	search := c.Query("search")

	var schools []domain.School
	var err error

	if search != "" {
		schools, err = h.repo.Search(c.Context(), search, limit)
	} else {
		schools, err = h.repo.List(c.Context(), regionID, limit, offset)
	}

	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "DB_ERROR", err.Error())
	}

	return response.Success(c, schools)
}
