package database

import (
	"context"
	"time"

	"apihorpug/config"
	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
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

// SeedAdmin ensures a full-access "Admin" role and a matching admin user
// exist, so the system can always be logged into after a fresh migration.
// It is idempotent: existing roles/users are left untouched.
func SeedAdmin(db *pgxpool.Pool, cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var roleID uuid.UUID
	if err := db.QueryRow(ctx, `
		INSERT INTO roles (id, name, description, is_active)
		VALUES ($1, 'Admin', 'Full system access', TRUE)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, uuid.New()).Scan(&roleID); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
		SELECT $1, m.id, p.id FROM menus m CROSS JOIN permissions p
		ON CONFLICT (role_id, menu_id, permission_id) DO NOTHING
	`, roleID); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `
		INSERT INTO users (id, username, email, password, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, TRUE)
		ON CONFLICT (username) DO NOTHING
	`, uuid.New(), cfg.AdminUsername, cfg.AdminEmail, string(hashedPassword), roleID); err != nil {
		return err
	}

	return nil
}
