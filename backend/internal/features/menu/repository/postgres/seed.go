package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedMenus(db *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	menus := []struct {
		Name        string
		Path        string
		Description string
		IsActive    bool
	}{
		{Name: "Users", Path: "/users", Description: "User management", IsActive: true},
		{Name: "Roles", Path: "/roles", Description: "Role and access management", IsActive: true},
		{Name: "Permissions", Path: "/permissions", Description: "Permission catalog", IsActive: true},
		{Name: "Dormitories", Path: "/dormitories", Description: "Dormitory management", IsActive: true},
		{Name: "Room Types", Path: "/room-types", Description: "Room type management", IsActive: true},
		{Name: "Rooms", Path: "/rooms", Description: "Room management", IsActive: true},
		{Name: "Tenants", Path: "/tenants", Description: "Tenant management", IsActive: true},
		{Name: "Contracts", Path: "/contracts", Description: "Lease contract management", IsActive: true},
		{Name: "Activity Logs", Path: "/activity-logs", Description: "Activity log records", IsActive: true},
	}

	for _, menu := range menus {
		if _, err := db.Exec(ctx, `
			INSERT INTO menus (id, name, path, description, is_active)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (path)
			DO UPDATE SET
				name = EXCLUDED.name,
				description = EXCLUDED.description,
				is_active = EXCLUDED.is_active,
				updated_at = NOW()
		`, uuid.New(), menu.Name, menu.Path, menu.Description, menu.IsActive); err != nil {
			return err
		}
	}

	return nil
}
