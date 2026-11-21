package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
)

type LimitsChecker interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
}

// CheckTokenLimit middleware checks if user has enough tokens
func CheckTokenLimit(repo domain.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uuid.UUID)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
		}

		// Get user with subscription info
		user, err := repo.GetByID(c.Context(), userID)
		if err != nil || user == nil {
			return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found")
		}

		// Check if subscription is active (not expired)
		if user.SubscriptionExpiresAt != nil && user.SubscriptionExpiresAt.Before(time.Now()) {
			return response.Error(c, fiber.StatusForbidden, "SUBSCRIPTION_EXPIRED", "Your subscription has expired. Please renew.", fiber.Map{
				"plan":       user.SubscriptionPlan,
				"expired_at": user.SubscriptionExpiresAt,
			})
		}

		// Ultra plan has unlimited tokens
		if user.SubscriptionPlan == domain.PlanUltra {
			return c.Next()
		}

		// Check token limit
		// Estimate: each AI request consumes ~10-50 tokens (will be calculated after generation)
		// For now, just check if user has ANY tokens left
		if user.TokensUsed >= user.TokensLimit {
			return response.Error(c, fiber.StatusForbidden, "TOKEN_LIMIT_EXCEEDED", "You have reached your token limit. Please upgrade your plan.", fiber.Map{
				"current_plan":  user.SubscriptionPlan,
				"tokens_used":   user.TokensUsed,
				"tokens_limit":  user.TokensLimit,
				"upgrade_plans": []string{"basic", "ultra"},
			})
		}

		// Store user in context for later use
		c.Locals("user", user)
		return c.Next()
	}
}

// CheckStudentLimit middleware checks if user can add more students
func CheckStudentLimit(repo domain.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(uuid.UUID)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "User ID not found")
		}

		user, err := repo.GetByID(c.Context(), userID)
		if err != nil || user == nil {
			return response.Error(c, fiber.StatusNotFound, "USER_NOT_FOUND", "User not found")
		}

		// Ultra plan has unlimited students
		if user.SubscriptionPlan == domain.PlanUltra {
			return c.Next()
		}

		// Check student limit
		if user.StudentsCount >= user.StudentsLimit {
			return response.Error(c, fiber.StatusForbidden, "STUDENT_LIMIT_EXCEEDED", "You have reached your student limit. Please upgrade your plan.", fiber.Map{
				"current_plan":     user.SubscriptionPlan,
				"students_count":   user.StudentsCount,
				"students_limit":   user.StudentsLimit,
				"upgrade_plans":    []string{"basic", "ultra"},
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
