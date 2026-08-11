package middleware

import (
	"context"
	"strings"
	"time"

	permissiondomain "apihorpug/internal/features/permission/domain"
	"apihorpug/internal/http/apierror"
	platformjwt "apihorpug/internal/platform/jwt"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	localsUserID   = "user_id"
	localsRoleID   = "role_id"
	localsUsername = "username"
)

func RequireAuth(secret string) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return apierror.Unauthorized("missing or invalid authorization header")
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		claims, err := platformjwt.Parse(secret, tokenString)
		if err != nil {
			return apierror.Unauthorized("invalid or expired token")
		}

		c.Locals(localsUserID, claims.UserID)
		c.Locals(localsRoleID, claims.RoleID)
		c.Locals(localsUsername, claims.Username)

		return c.Next()
	}
}

func RequirePermission(db *pgxpool.Pool, menuPath string, action permissiondomain.Action) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, ok := UserID(c)
		if !ok {
			return apierror.Unauthorized("authentication required")
		}

		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		var exists int
		err := db.QueryRow(ctx, `
			SELECT 1
			FROM users u
			JOIN role_menu_permissions rmp ON rmp.role_id = u.role_id
			JOIN menus m ON m.id = rmp.menu_id
			JOIN permissions p ON p.id = rmp.permission_id
			WHERE u.id = $1 AND u.is_active = TRUE AND m.path = $2 AND p.name = $3
			LIMIT 1
		`, userID, menuPath, string(action)).Scan(&exists)
		if err != nil {
			if err == pgx.ErrNoRows {
				return apierror.Forbidden("insufficient permission")
			}
			return apierror.Internal("failed to verify permission")
		}

		return c.Next()
	}
}

func UserID(c fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsUserID).(uuid.UUID)
	return id, ok
}

func RoleID(c fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(localsRoleID).(uuid.UUID)
	return id, ok
}

func Username(c fiber.Ctx) (string, bool) {
	username, ok := c.Locals(localsUsername).(string)
	return username, ok
}
