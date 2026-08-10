package middleware

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/dormitory/usecase"

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
			accessible, err := dormitories.ListAccessible(c.Context(), userID, roleName)
			if err != nil {
				return err
			}
			if len(accessible) == 0 {
				return apierror.Forbidden("no accessible dormitory")
			}
			dormitoryID = accessible[0].ID
		} else {
			ok, err := dormitories.CheckAccess(c.Context(), userID, roleName, dormitoryID)
			if err != nil {
				return err
			}
			if !ok {
				return apierror.Forbidden("no access to this dormitory")
			}
		}

		c.Locals("dormitory_id", dormitoryID)
		return c.Next()
	}
}
