package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
			created_by UUID,
			updated_by UUID,
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
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT users_role_id_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON UPDATE CASCADE ON DELETE RESTRICT,
			CONSTRAINT users_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT users_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
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
		`CREATE TABLE IF NOT EXISTS dormitories (
			id UUID PRIMARY KEY,
			name VARCHAR(150) NOT NULL,
			address VARCHAR(255) DEFAULT '',
			phone VARCHAR(30) DEFAULT '',
			description VARCHAR(255) DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_dormitories (
			user_id UUID NOT NULL,
			dormitory_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (user_id, dormitory_id),
			CONSTRAINT user_dormitories_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT user_dormitories_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS activity_logs (
			id UUID PRIMARY KEY,
			user_id UUID,
			action VARCHAR(50) NOT NULL,
			entity_type VARCHAR(80) NOT NULL,
			entity_id UUID,
			description VARCHAR(255) DEFAULT '',
			ip_address VARCHAR(45) DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT activity_logs_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_entity ON activity_logs(entity_type, entity_id)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	auditColumns := []string{"users", "roles", "dormitories"}
	for _, table := range auditColumns {
		if _, err := db.Exec(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_by UUID`, table)); err != nil {
			return err
		}
		if _, err := db.Exec(ctx, fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS updated_by UUID`, table)); err != nil {
			return err
		}
	}

	if _, err := db.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_created_by_fkey') THEN
				ALTER TABLE users ADD CONSTRAINT users_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'users_updated_by_fkey') THEN
				ALTER TABLE users ADD CONSTRAINT users_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'roles_created_by_fkey') THEN
				ALTER TABLE roles ADD CONSTRAINT roles_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'roles_updated_by_fkey') THEN
				ALTER TABLE roles ADD CONSTRAINT roles_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'dormitories_created_by_fkey') THEN
				ALTER TABLE dormitories ADD CONSTRAINT dormitories_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'dormitories_updated_by_fkey') THEN
				ALTER TABLE dormitories ADD CONSTRAINT dormitories_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
		END
		$$;
	`); err != nil {
		return err
	}

	return nil
}
