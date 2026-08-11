package database

import (
	"context"
	"fmt"
	"time"

	"apihorpug/config"
	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgres(cfg config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=Asia/Bangkok",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUsername,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func AutoMigrate(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS menus (
			id UUID PRIMARY KEY,
			name VARCHAR(120) NOT NULL,
			path VARCHAR(255) UNIQUE NOT NULL,
			description VARCHAR(255) DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id UUID PRIMARY KEY,
			name VARCHAR(120) UNIQUE NOT NULL,
			description VARCHAR(255) DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY,
			name VARCHAR(120) UNIQUE NOT NULL,
			description VARCHAR(255) DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(80) UNIQUE NOT NULL,
			email VARCHAR(180) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role_id UUID NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON UPDATE CASCADE ON DELETE RESTRICT
		)`,
		`CREATE TABLE IF NOT EXISTS role_menu_permissions (
			role_id UUID NOT NULL,
			menu_id UUID NOT NULL,
			permission_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (role_id, menu_id, permission_id),
			CONSTRAINT rmp_role_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			CONSTRAINT rmp_menu_fkey FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE,
			CONSTRAINT rmp_permission_fkey FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	return nil
}

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

