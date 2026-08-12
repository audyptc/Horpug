package postgres

import (
	"context"
	"errors"

	permissiondomain "apihorpug/internal/features/permission/domain"
	permissionusecase "apihorpug/internal/features/permission/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
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
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM permissions`).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]permissiondomain.Permission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM permissions
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]permissiondomain.Permission, 0)
	for rows.Next() {
		var permission permissiondomain.Permission
		if err := rows.Scan(&permission.ID, &permission.Name, &permission.Description, &permission.CreatedAt, &permission.UpdatedAt); err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *Repository) Create(ctx context.Context, input permissionusecase.CreateInput) (permissiondomain.Permission, error) {
	permission := permissiondomain.Permission{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO permissions (id, name, description)
		VALUES ($1, $2, $3)
		RETURNING created_at, updated_at
	`, permission.ID, permission.Name, permission.Description).Scan(&permission.CreatedAt, &permission.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return permissiondomain.Permission{}, permissiondomain.ErrPermissionNameExists
		}
		return permissiondomain.Permission{}, err
	}

	return permission, nil
}
