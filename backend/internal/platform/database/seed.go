package database

import (
	"context"
	"time"

	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedPermissions(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	descriptions := map[permissiondomain.Action]string{
		permissiondomain.ActionCreate: "Create records",
		permissiondomain.ActionRead:   "View records",
		permissiondomain.ActionUpdate: "Update records",
		permissiondomain.ActionDelete: "Delete records",
	}

	names := make([]string, 0, len(permissiondomain.Actions))
	for _, action := range permissiondomain.Actions {
		names = append(names, string(action))

		if _, err := db.Exec(ctx, `
			INSERT INTO permissions (id, name, description)
			VALUES ($1, $2, $3)
			ON CONFLICT (name)
			DO UPDATE SET description = EXCLUDED.description, updated_at = NOW()
		`, uuid.New(), string(action), descriptions[action]); err != nil {
			return err
		}
	}

	// Remove leftover resource-scoped permissions (e.g. "users.read") now that
	// permissions are central CRUD actions bound to menus via role_menu_permissions.
	if _, err := db.Exec(ctx, `DELETE FROM permissions WHERE NOT (name = ANY($1))`, names); err != nil {
		return err
	}

	return nil
}
