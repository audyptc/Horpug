package postgres

import (
	"context"

	menudomain "apihorpug/internal/features/menu/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]menudomain.Menu, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, path, description, is_active, created_at, updated_at
		FROM menus
		ORDER BY path ASC
	`)
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
