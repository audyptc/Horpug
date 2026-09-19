package postgres

import (
	"context"

	menudomain "apihorpug/internal/features/menu/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM menus`).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]menudomain.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, path, description, is_active, created_at, updated_at
		FROM menus
		ORDER BY path ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus := make([]menudomain.Menu, 0)
	for rows.Next() {
		var menu menudomain.Menu
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Path, &menu.Description, &menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return menus, nil
}

// ListForRole returns the active menus a role has the given permission
// (action) on, e.g. the menus a user should see in navigation.
func (r *Repository) ListForRole(ctx context.Context, roleID uuid.UUID, action string) ([]menudomain.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT m.id, m.name, m.path, m.description, m.is_active, m.created_at, m.updated_at
		FROM menus m
		JOIN role_menu_permissions rmp ON rmp.menu_id = m.id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE rmp.role_id = $1 AND p.name = $2 AND m.is_active = TRUE
		ORDER BY m.path ASC
	`, roleID, action)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menus := make([]menudomain.Menu, 0)
	for rows.Next() {
		var menu menudomain.Menu
		if err := rows.Scan(&menu.ID, &menu.Name, &menu.Path, &menu.Description, &menu.IsActive, &menu.CreatedAt, &menu.UpdatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, menu)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return menus, nil
}
