package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	coredomain "apigofiberhorpug/internal/domain"
	roledomain "apigofiberhorpug/internal/feature/role/domain"
	"apigofiberhorpug/internal/feature/user/domain"

	"github.com/jackc/pgx/v5"
)

type UserRepo struct {
	db *database.DB
}

func NewUserRepo(db *database.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, full_name, email, password, is_active, created_at, updated_at
		FROM users WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", coredomain.ErrNotFound)
	}
	return u, err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, full_name, email, password, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND deleted_at IS NULL`, email).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found: %w", coredomain.ErrNotFound)
	}
	return u, err
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, full_name, email, is_active, created_at, updated_at
		FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if users == nil {
		users = []*domain.User{}
	}
	return users, rows.Err()
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO users (id, full_name, email, password, is_active)
		VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.FullName, user.Email, user.Password, user.IsActive)
	return err
}

func (r *UserRepo) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE users SET full_name = $2, is_active = $3, password = $4, updated_at = NOW()
		WHERE id = $1`,
		user.ID, user.FullName, user.IsActive, user.Password)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}

func (r *UserRepo) AssignRole(ctx context.Context, userID string, roleID string) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id)
		DO UPDATE SET role_id = EXCLUDED.role_id, created_at = NOW()`,
		userID, roleID)
	return err
}

func (r *UserRepo) GetRole(ctx context.Context, userID string) (*roledomain.Role, error) {
	role := &roledomain.Role{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT r.id, r.name, r.description, r.is_active, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1 AND r.is_active = TRUE`, userID).
		Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *UserRepo) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT TRIM(BOTH '/' FROM m.path) || '.' || p.name AS permission_name
		FROM role_menu_permissions rmp
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		JOIN user_roles ur ON ur.role_id = rmp.role_id
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1 AND r.is_active = TRUE
		ORDER BY permission_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}
	if permissions == nil {
		permissions = []string{}
	}
	return permissions, rows.Err()
}

func (r *UserRepo) GetDormitoryPermissions(ctx context.Context, userID string) (map[string][]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT udr.dormitory_id, TRIM(BOTH '/' FROM m.path) || '.' || p.name AS permission_name
		FROM user_dormitory_roles udr
		JOIN roles r ON r.id = udr.role_id
		JOIN role_menu_permissions rmp ON rmp.role_id = r.id
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE udr.user_id = $1 AND r.is_active = TRUE
		ORDER BY udr.dormitory_id, permission_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissionsByDormitory := map[string][]string{}
	for rows.Next() {
		var dormitoryID string
		var permissionName string
		if err := rows.Scan(&dormitoryID, &permissionName); err != nil {
			return nil, err
		}
		permissionsByDormitory[dormitoryID] = append(permissionsByDormitory[dormitoryID], permissionName)
	}
	return permissionsByDormitory, rows.Err()
}
