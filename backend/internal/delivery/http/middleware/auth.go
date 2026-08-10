package middleware

import (
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/auth/usecase"

	"github.com/gofiber/fiber/v3"
)

func RequireAuth(auth *usecase.AuthUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return apierror.Unauthorized("missing or invalid authorization header")
		}

		claims, err := auth.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			return apierror.Unauthorized("invalid or expired token")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role_name", claims.RoleName)
		c.Locals("permissions", claims.GlobalPermissions)
		c.Locals("global_permissions", claims.GlobalPermissions)
		c.Locals("dormitory_permissions", claims.DormitoryPermissions)
		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if roleName, _ := c.Locals("role_name").(string); strings.EqualFold(roleName, "admin") {
			return c.Next()
		}
		perms, _ := c.Locals("permissions").([]string)
		for _, p := range perms {
			if p == permission {
				return c.Next()
			}
		}
		return apierror.Forbidden("insufficient permissions")
	}
}
