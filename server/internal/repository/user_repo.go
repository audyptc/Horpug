package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

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
		FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, full_name, email, password, is_active, created_at, updated_at
		FROM users WHERE email = $1`, email).
		Scan(&u.ID, &u.FullName, &u.Email, &u.Password, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, full_name, email, is_active, created_at, updated_at
		FROM users ORDER BY created_at DESC
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
	return users, nil
}

func (r *UserRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
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
		UPDATE users SET full_name = $2, is_active = $3, updated_at = NOW()
		WHERE id = $1`,
		user.ID, user.FullName, user.IsActive)
	return err
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *UserRepo) AssignRoles(ctx context.Context, userID string, roleIDs []string) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			userID, roleID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *UserRepo) GetRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	if roles == nil {
		roles = []domain.Role{}
	}
	return roles, nil
}

func (r *UserRepo) GetPermissions(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT p.name
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
		ORDER BY p.name`, userID)
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
	return permissions, nil
}
