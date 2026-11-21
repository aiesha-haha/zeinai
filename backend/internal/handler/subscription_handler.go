package handler

import (
	"time"

	"github.com/go-playground/validator/v10"
	"zeinai-backend/internal/pkg/validation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/service"
)

type SubscriptionHandler struct {
	service   *service.SubscriptionService
	userService *service.UserService
	validator *validator.Validate
}

func NewSubscriptionHandler(s *service.SubscriptionService, us *service.UserService) *SubscriptionHandler {
	return &SubscriptionHandler{
		service:   s,
		userService: us,
		validator: validation.NewValidator(),
	}
}

// GetSubscription godoc
// @Summary Get current subscription details
// @Description Get detailed information about user's subscription
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=domain.SubscriptionResponse}
// @Failure 401 {object} response.APIResponse
// @Router /subscriptions/current [get]
func (h *SubscriptionHandler) GetSubscription(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	}

	// Calculate days remaining
	var daysRemaining *int
	var expiresAtStr *string
	isActive := true
	if profile.Subscription.Plan != domain.PlanFree && profile.Subscription.ExpiresAt != nil {
		expiresAtStr = new(string)
		*expiresAtStr = profile.Subscription.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		
		days := int(time.Until(*profile.Subscription.ExpiresAt).Hours() / 24)
		daysRemaining = &days
		
		if days < 0 {
			isActive = false
		}
	} else if profile.Subscription.Plan == domain.PlanFree {
		isActive = true
		expiresAtStr = nil
		daysRemaining = nil
	}

	limitsResetAtStr := new(string)
	if profile.Subscription.LimitsResetAt != nil {
		*limitsResetAtStr = profile.Subscription.LimitsResetAt.Format("2006-01-02T15:04:05Z07:00")
	}

	subscriptionResp := domain.SubscriptionResponse{
		Plan:              profile.Subscription.Plan,
		TokensUsed:        profile.Subscription.TokensUsed,
		TokensLimit:       profile.Subscription.TokensLimit,
		TokensRemaining:   profile.Subscription.TokensRemaining,
		StudentsCount:     profile.Subscription.StudentsCount,
		StudentsLimit:     profile.Subscription.StudentsLimit,
		StudentsRemaining: profile.Subscription.StudentsRemaining,
		StartedAt:         profile.Subscription.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:         expiresAtStr,
		LimitsResetAt:     limitsResetAtStr,
		IsActive:          isActive,
		DaysRemaining:     daysRemaining,
	}

	return response.Success(c, subscriptionResp)
}

// UpgradeSubscription godoc
// @Summary Upgrade subscription plan
// @Description Upgrade user's subscription to a higher plan
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpgradeSubscriptionRequest true "Upgrade Request"
// @Success 200 {object} response.APIResponse{data=domain.SubscriptionResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /subscriptions/upgrade [post]
func (h *SubscriptionHandler) UpgradeSubscription(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.UpgradeSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// Validate plan upgrade logic
	if req.Plan == domain.PlanFree {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_UPGRADE", "Cannot downgrade to free plan via API")
	}

	// TODO: Integrate payment gateway here
	// For now, we'll just upgrade without payment verification

	err := h.service.UpgradePlan(c.Context(), userID, req.Plan, req.DurationMonths)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "UPGRADE_FAILED", err.Error())
	}

	// Get updated subscription
	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "PROFILE_FETCH_FAILED", err.Error())
	}

	var expiresAtStr *string
	if profile.Subscription.ExpiresAt != nil {
		expiresAtStr = new(string)
		*expiresAtStr = profile.Subscription.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	subscriptionResp := domain.SubscriptionResponse{
		Plan:              profile.Subscription.Plan,
		TokensUsed:        profile.Subscription.TokensUsed,
		TokensLimit:       profile.Subscription.TokensLimit,
		TokensRemaining:   profile.Subscription.TokensRemaining,
		StudentsCount:     profile.Subscription.StudentsCount,
		StudentsLimit:     profile.Subscription.StudentsLimit,
		StudentsRemaining: profile.Subscription.StudentsRemaining,
		StartedAt:         profile.Subscription.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:         expiresAtStr,
		IsActive:          true,
	}

	return response.Success(c, subscriptionResp)
}

// RenewSubscription godoc
// @Summary Renew current subscription
// @Description Extend current subscription by specified months
// @Tags Subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.UpgradeSubscriptionRequest true "Renew Request"
// @Success 200 {object} response.APIResponse{data=domain.SubscriptionResponse}
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /subscriptions/renew [post]
func (h *SubscriptionHandler) RenewSubscription(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
	}

	var req domain.UpgradeSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_BODY", err.Error())
	}

	// Get current profile to check plan
	profile, err := h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", err.Error())
	}

	// Use current plan for renewal
	currentPlan := profile.Subscription.Plan
	if currentPlan == domain.PlanFree {
		return response.Error(c, fiber.StatusBadRequest, "INVALID_RENEWAL", "Free plan cannot be renewed. Please upgrade first.")
	}

	// TODO: Integrate payment gateway

	err = h.service.UpgradePlan(c.Context(), userID, currentPlan, req.DurationMonths)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "RENEWAL_FAILED", err.Error())
	}

	// Get updated subscription
	profile, err = h.userService.GetProfile(c.Context(), userID)
	if err != nil {
		return response.Error(c, fiber.StatusInternalServerError, "PROFILE_FETCH_FAILED", err.Error())
	}

	var expiresAtStr *string
	if profile.Subscription.ExpiresAt != nil {
		expiresAtStr = new(string)
		*expiresAtStr = profile.Subscription.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
	}

	subscriptionResp := domain.SubscriptionResponse{
		Plan:              profile.Subscription.Plan,
		TokensUsed:        profile.Subscription.TokensUsed,
		TokensLimit:       profile.Subscription.TokensLimit,
		TokensRemaining:   profile.Subscription.TokensRemaining,
		StudentsCount:     profile.Subscription.StudentsCount,
		StudentsLimit:     profile.Subscription.StudentsLimit,
		StudentsRemaining: profile.Subscription.StudentsRemaining,
		StartedAt:         profile.Subscription.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
		ExpiresAt:         expiresAtStr,
		IsActive:          true,
	}

	return response.Success(c, subscriptionResp)
}
