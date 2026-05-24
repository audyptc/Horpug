package middleware

import (
	"strings"

	"apigofiberhorpug/internal/usecase"

	"github.com/gofiber/fiber/v3"
)

func RequireAuth(auth *usecase.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "missing or invalid authorization header",
			})
		}

		claims, err := auth.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "invalid or expired token",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("permissions", claims.Permissions)
		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		perms, _ := c.Locals("permissions").([]string)
		for _, p := range perms {
			if p == permission {
				return c.Next()
			}
		}
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "insufficient permissions",
		})
	}
}
