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
			full_dormitory_access BOOLEAN NOT NULL DEFAULT FALSE,
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
		`CREATE TABLE IF NOT EXISTS role_dormitories (
			role_id UUID NOT NULL,
			dormitory_id UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (role_id, dormitory_id),
			CONSTRAINT role_dormitories_role_fkey FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
			CONSTRAINT role_dormitories_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS room_types (
			id UUID PRIMARY KEY,
			dormitory_id UUID NOT NULL,
			name VARCHAR(120) NOT NULL,
			description VARCHAR(255) DEFAULT '',
			price NUMERIC(10,2) NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT room_types_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE,
			CONSTRAINT room_types_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT room_types_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT uq_room_types_dormitory_name UNIQUE (dormitory_id, name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_room_types_dormitory_id ON room_types(dormitory_id)`,
		`CREATE TABLE IF NOT EXISTS rooms (
			id UUID PRIMARY KEY,
			dormitory_id UUID NOT NULL,
			room_type_id UUID NOT NULL,
			room_number VARCHAR(20) NOT NULL,
			floor INTEGER NOT NULL DEFAULT 1,
			status VARCHAR(20) NOT NULL DEFAULT 'available',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT rooms_dormitory_fkey FOREIGN KEY (dormitory_id) REFERENCES dormitories(id) ON DELETE CASCADE,
			CONSTRAINT rooms_room_type_fkey FOREIGN KEY (room_type_id) REFERENCES room_types(id) ON DELETE RESTRICT,
			CONSTRAINT rooms_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT rooms_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT uq_rooms_dormitory_room_number UNIQUE (dormitory_id, room_number)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_dormitory_id ON rooms(dormitory_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rooms_room_type_id ON rooms(room_type_id)`,
		`CREATE TABLE IF NOT EXISTS tenants (
			id UUID PRIMARY KEY,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone VARCHAR(30) DEFAULT '',
			id_card VARCHAR(20) DEFAULT '',
			email VARCHAR(180) DEFAULT '',
			emergency_contact VARCHAR(150) DEFAULT '',
			note VARCHAR(255) DEFAULT '',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT tenants_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT tenants_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_tenants_id_card ON tenants(id_card) WHERE id_card <> ''`,
		`CREATE TABLE IF NOT EXISTS contracts (
			id UUID PRIMARY KEY,
			tenant_id UUID NOT NULL,
			room_id UUID NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			rent_price NUMERIC(10,2) NOT NULL DEFAULT 0,
			deposit NUMERIC(10,2) NOT NULL DEFAULT 0,
			num_occupants INTEGER NOT NULL DEFAULT 1,
			status VARCHAR(20) NOT NULL DEFAULT 'active',
			note VARCHAR(255) DEFAULT '',
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT contracts_tenant_fkey FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT,
			CONSTRAINT contracts_room_fkey FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE RESTRICT,
			CONSTRAINT contracts_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT contracts_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_tenant_id ON contracts(tenant_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contracts_room_id ON contracts(room_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_contracts_active_room ON contracts(room_id) WHERE status = 'active'`,
		`CREATE TABLE IF NOT EXISTS electricity_meters (
			id UUID PRIMARY KEY,
			room_id UUID NOT NULL,
			billing_method VARCHAR(20) NOT NULL DEFAULT 'metered',
			reading_date DATE NOT NULL,
			previous_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			current_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			unit_used NUMERIC(10,2) GENERATED ALWAYS AS (current_unit - previous_unit) STORED,
			price_per_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			flat_amount NUMERIC(10,2),
			total_amount NUMERIC(10,2) GENERATED ALWAYS AS (
				CASE WHEN billing_method = 'flat' THEN COALESCE(flat_amount, 0)
				ELSE (current_unit - previous_unit) * price_per_unit
				END
			) STORED,
			note VARCHAR(255) DEFAULT '',
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT electricity_meters_room_fkey FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
			CONSTRAINT electricity_meters_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT electricity_meters_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT uq_electricity_meters_room_date UNIQUE (room_id, reading_date),
			CONSTRAINT chk_electricity_meters_units CHECK (previous_unit >= 0 AND current_unit >= previous_unit),
			CONSTRAINT chk_electricity_meters_billing_method CHECK (billing_method IN ('metered', 'flat')),
			CONSTRAINT chk_electricity_meters_flat_amount CHECK (billing_method <> 'flat' OR flat_amount IS NOT NULL),
			CONSTRAINT chk_electricity_meters_flat_amount_nonneg CHECK (flat_amount IS NULL OR flat_amount >= 0)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_electricity_meters_room_id ON electricity_meters(room_id)`,
		`CREATE TABLE IF NOT EXISTS water_meters (
			id UUID PRIMARY KEY,
			room_id UUID NOT NULL,
			billing_method VARCHAR(20) NOT NULL DEFAULT 'metered',
			reading_date DATE NOT NULL,
			previous_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			current_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			unit_used NUMERIC(10,2) GENERATED ALWAYS AS (current_unit - previous_unit) STORED,
			price_per_unit NUMERIC(10,2) NOT NULL DEFAULT 0,
			flat_amount NUMERIC(10,2),
			total_amount NUMERIC(10,2) GENERATED ALWAYS AS (
				CASE WHEN billing_method = 'flat' THEN COALESCE(flat_amount, 0)
				ELSE (current_unit - previous_unit) * price_per_unit
				END
			) STORED,
			note VARCHAR(255) DEFAULT '',
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT water_meters_room_fkey FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
			CONSTRAINT water_meters_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT water_meters_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT uq_water_meters_room_date UNIQUE (room_id, reading_date),
			CONSTRAINT chk_water_meters_units CHECK (previous_unit >= 0 AND current_unit >= previous_unit),
			CONSTRAINT chk_water_meters_billing_method CHECK (billing_method IN ('metered', 'flat')),
			CONSTRAINT chk_water_meters_flat_amount CHECK (billing_method <> 'flat' OR flat_amount IS NOT NULL),
			CONSTRAINT chk_water_meters_flat_amount_nonneg CHECK (flat_amount IS NULL OR flat_amount >= 0)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_water_meters_room_id ON water_meters(room_id)`,
		`CREATE TABLE IF NOT EXISTS invoices (
			id UUID PRIMARY KEY,
			contract_id UUID NOT NULL,
			period_year INTEGER NOT NULL,
			period_month INTEGER NOT NULL,
			issue_date DATE NOT NULL,
			due_date DATE NOT NULL,
			total_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
			status VARCHAR(20) NOT NULL DEFAULT 'unpaid',
			paid_at TIMESTAMPTZ,
			note VARCHAR(255) DEFAULT '',
			created_by UUID,
			updated_by UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT invoices_contract_fkey FOREIGN KEY (contract_id) REFERENCES contracts(id) ON DELETE RESTRICT,
			CONSTRAINT invoices_created_by_fkey FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT invoices_updated_by_fkey FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT uq_invoices_contract_period UNIQUE (contract_id, period_year, period_month),
			CONSTRAINT chk_invoices_period_month CHECK (period_month BETWEEN 1 AND 12),
			CONSTRAINT chk_invoices_status CHECK (status IN ('unpaid', 'paid', 'overdue', 'cancelled'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_contract_id ON invoices(contract_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_status ON invoices(status)`,
		`CREATE TABLE IF NOT EXISTS invoice_items (
			id UUID PRIMARY KEY,
			invoice_id UUID NOT NULL,
			item_type VARCHAR(20) NOT NULL,
			description VARCHAR(255) NOT NULL DEFAULT '',
			reference_id UUID,
			amount NUMERIC(10,2) NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT invoice_items_invoice_fkey FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			CONSTRAINT chk_invoice_items_type CHECK (item_type IN ('rent', 'electricity', 'water', 'other'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_invoice_items_invoice_id ON invoice_items(invoice_id)`,
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
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			token_hash VARCHAR(64) UNIQUE NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			revoked_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			CONSTRAINT refresh_tokens_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	compatibilityStatements := []string{
		`ALTER TABLE menus ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT ''`,
		`ALTER TABLE menus ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE menus ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE menus ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_menus_path ON menus(path)`,

		`ALTER TABLE permissions ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT ''`,
		`ALTER TABLE permissions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE permissions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_permissions_name ON permissions(name)`,

		`ALTER TABLE room_types ADD COLUMN IF NOT EXISTS price NUMERIC(10,2) NOT NULL DEFAULT 0`,

		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS description VARCHAR(255) DEFAULT ''`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS full_dormitory_access BOOLEAN NOT NULL DEFAULT FALSE`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE roles ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_roles_name ON roles(name)`,

		`ALTER TABLE electricity_meters ADD COLUMN IF NOT EXISTS billing_method VARCHAR(20) NOT NULL DEFAULT 'metered'`,
		`ALTER TABLE electricity_meters ADD COLUMN IF NOT EXISTS flat_amount NUMERIC(10,2)`,

		`ALTER TABLE users ADD COLUMN IF NOT EXISTS username VARCHAR(80)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS email VARCHAR(180)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password VARCHAR(255)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role_id UUID`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_users_username ON users(username)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_users_email ON users(email)`,
	}

	for _, stmt := range compatibilityStatements {
		if _, err := db.Exec(ctx, stmt); err != nil {
			return err
		}
	}

	// electricity_meters.total_amount originally only supported the metered
	// formula; re-generate it to also support flat-rate billing once
	// billing_method/flat_amount exist. Only rewrites tables still on the old
	// expression, so re-running this is a no-op once upgraded.
	if _, err := db.Exec(ctx, `
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'electricity_meters'
				AND column_name = 'total_amount'
				AND generation_expression NOT ILIKE '%billing_method%'
			) THEN
				ALTER TABLE electricity_meters DROP COLUMN total_amount;
				ALTER TABLE electricity_meters ADD COLUMN total_amount NUMERIC(10,2) GENERATED ALWAYS AS (
					CASE WHEN billing_method = 'flat' THEN COALESCE(flat_amount, 0)
					ELSE (current_unit - previous_unit) * price_per_unit
					END
				) STORED;
			END IF;

			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_electricity_meters_billing_method') THEN
				ALTER TABLE electricity_meters ADD CONSTRAINT chk_electricity_meters_billing_method CHECK (billing_method IN ('metered', 'flat'));
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_electricity_meters_flat_amount') THEN
				ALTER TABLE electricity_meters ADD CONSTRAINT chk_electricity_meters_flat_amount CHECK (billing_method <> 'flat' OR flat_amount IS NOT NULL);
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_electricity_meters_flat_amount_nonneg') THEN
				ALTER TABLE electricity_meters ADD CONSTRAINT chk_electricity_meters_flat_amount_nonneg CHECK (flat_amount IS NULL OR flat_amount >= 0);
			END IF;
		END
		$$;
	`); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'full_name'
			) THEN
				ALTER TABLE users ALTER COLUMN full_name SET DEFAULT '';
			END IF;
		END
		$$;
	`); err != nil {
		return err
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

	if _, err := db.Exec(ctx, `ALTER TABLE activity_logs ADD COLUMN IF NOT EXISTS user_id UUID`); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id)`); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_activity_logs_entity ON activity_logs(entity_type, entity_id)`); err != nil {
		return err
	}
	if _, err := db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at DESC)`); err != nil {
		return err
	}

	if _, err := db.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'activity_logs_user_fkey') THEN
				ALTER TABLE activity_logs ADD CONSTRAINT activity_logs_user_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
			END IF;
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
