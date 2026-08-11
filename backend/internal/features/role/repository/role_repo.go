package repository

import (
	"context"
	"fmt"
	"time"

	permissiondomain "apigofiberhorpug/internal/features/permission/domain"
	"apigofiberhorpug/internal/features/role/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

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
		SELECT id, name, description, is_active, created_at, updated_at
		FROM roles WHERE id = $1`, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("role not found: %w", coredomain.ErrNotFound)
	}
	return role, err
}

func (r *RoleRepo) List(ctx context.Context, limit, offset int) ([]*domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM roles ORDER BY name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if roles == nil {
		roles = []*domain.Role{}
	}
	return roles, rows.Err()
}

func (r *RoleRepo) ListActive(ctx context.Context) ([]*domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM roles WHERE is_active = TRUE ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt); err != nil {
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
		INSERT INTO roles (id, name, description, is_active) VALUES ($1, $2, $3, $4)`,
		role.ID, role.Name, role.Description, role.IsActive)
	return err
}

func (r *RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE roles SET name = $2, description = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1`,
		role.ID, role.Name, role.Description, role.IsActive)
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

func (r *RoleRepo) GetMenuPermissions(ctx context.Context, roleID string) ([]domain.RoleMenuPermission, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT m.id, m.name, p.id, p.name, p.description, p.created_at, p.updated_at
		FROM role_menu_permissions rmp
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE rmp.role_id = $1
		ORDER BY m.name, p.name`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	menuMap := make(map[string]*domain.RoleMenuPermission)
	var menuOrder []string

	for rows.Next() {
		var menuID, menuName, permID, permName, permDesc string
		var permCreatedAt, permUpdatedAt time.Time
		if err := rows.Scan(&menuID, &menuName, &permID, &permName, &permDesc, &permCreatedAt, &permUpdatedAt); err != nil {
			return nil, err
		}
		if _, exists := menuMap[menuID]; !exists {
			menuMap[menuID] = &domain.RoleMenuPermission{
				MenuID:      menuID,
				MenuName:    menuName,
				Permissions: []permissiondomain.Permission{},
			}
			menuOrder = append(menuOrder, menuID)
		}
		menuMap[menuID].Permissions = append(menuMap[menuID].Permissions, permissiondomain.Permission{
			ID:          permID,
			Name:        permName,
			Description: permDesc,
			CreatedAt:   permCreatedAt,
			UpdatedAt:   permUpdatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make([]domain.RoleMenuPermission, 0, len(menuOrder))
	for _, id := range menuOrder {
		result = append(result, *menuMap[id])
	}
	return result, nil
}
