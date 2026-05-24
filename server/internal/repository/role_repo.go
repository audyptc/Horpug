package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type RoleRepo struct {
	db *database.DB
}

func NewRoleRepo(db *database.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) FindByID(ctx context.Context, id string) (*domain.Role, error) {
	role := &domain.Role{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM roles WHERE id = $1`, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("role not found")
	}
	return role, err
}

func (r *RoleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM roles ORDER BY name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if roles == nil {
		roles = []*domain.Role{}
	}
	return roles, rows.Err()
}

func (r *RoleRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM roles`).Scan(&total)
	return total, err
}

func (r *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO roles (id, name, description) VALUES ($1, $2, $3)`,
		role.ID, role.Name, role.Description)
	return err
}

func (r *RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE roles SET name = $2, description = $3, updated_at = NOW()
		WHERE id = $1`,
		role.ID, role.Name, role.Description)
	return err
}

func (r *RoleRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	return err
}

func (r *RoleRepo) AssignMenuPermissions(ctx context.Context, roleID string, items []domain.RoleMenuPermissionItem) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM role_menu_permissions WHERE role_id = $1`, roleID); err != nil {
		return err
	}
	for _, item := range items {
		for _, permID := range item.PermissionIDs {
			if _, err := tx.Exec(ctx,
				`INSERT INTO role_menu_permissions (role_id, menu_id, permission_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
				roleID, item.MenuID, permID); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *RoleRepo) GetPermissions(ctx context.Context, roleID string) ([]domain.Permission, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT
			m.id::text || ':' || p.id::text AS id,
			TRIM(BOTH '/' FROM m.path) || '.' || p.name AS name,
			m.name || ' - ' || p.description AS description,
			GREATEST(m.created_at, p.created_at) AS created_at,
			GREATEST(m.updated_at, p.updated_at) AS updated_at
		FROM role_menu_permissions rmp
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE rmp.role_id = $1
		ORDER BY name`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if perms == nil {
		perms = []domain.Permission{}
	}
	return perms, rows.Err()
}
