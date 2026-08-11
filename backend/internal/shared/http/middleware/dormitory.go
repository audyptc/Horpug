package middleware

import (
	"strings"

	"apigofiberhorpug/internal/features/dormitory/usecase"
	"apigofiberhorpug/internal/shared/http/apierror"

	"github.com/gofiber/fiber/v3"
)

// RequireDormitory resolves the current request's dormitory scope from the
// X-Dormitory-Id header, verifying the authenticated user has access to it.
// If the header is absent, it falls back to the user's first accessible
// dormitory so existing clients that don't yet send the header keep working.
func RequireDormitory(dormitories *usecase.DormitoryUseCase) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		roleName, _ := c.Locals("role_name").(string)
		dormitoryID := c.Get("X-Dormitory-Id")

		if dormitoryID == "" {
			return apierror.BadRequest("X-Dormitory-Id header is required")
		}

		ok, err := dormitories.CheckAccess(c.Context(), userID, roleName, dormitoryID)
		if err != nil {
			return err
		}
		if !ok {
			return apierror.Forbidden("no access to this dormitory")
		}

		dormitory, err := dormitories.GetByID(c.Context(), dormitoryID)
		if err != nil {
			return err
		}
		if !dormitory.IsActive {
			return apierror.Forbidden("dormitory is inactive")
		}

		if !strings.EqualFold(roleName, "admin") {
			permissionsByDormitory, _ := c.Locals("dormitory_permissions").(map[string][]string)
			permissions := []string{}
			if permissionsByDormitory != nil {
				permissions = permissionsByDormitory[dormitoryID]
			}
			c.Locals("permissions", permissions)
		}

		c.Locals("dormitory_id", dormitoryID)
		return c.Next()
	}
}
