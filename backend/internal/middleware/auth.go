package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"zeinai-backend/internal/config"
	"zeinai-backend/internal/domain"
	"zeinai-backend/internal/pkg/response"
	"zeinai-backend/internal/pkg/utils"
)

// Protected verifies the JWT token
func Protected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format")
		}

		token, err := jwt.ParseWithClaims(parts[1], &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid or expired token")
		}

		claims, ok := token.Claims.(*utils.Claims)
		if !ok {
			return response.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "Invalid token claims")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("role", claims.Role)
		return c.Next()
	}
}

// RBAC checks if the user has the required role
func RBAC(allowedRoles ...domain.Role) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(domain.Role)
		if !ok {
			return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "User role not found")
		}

		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return response.Error(c, fiber.StatusForbidden, "FORBIDDEN", "Insufficient permissions")
	}
}
