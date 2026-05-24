package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type PermissionRepo struct {
	db *database.DB
}

func NewPermissionRepo(db *database.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

func (r *PermissionRepo) List(ctx context.Context, limit, offset int) ([]*domain.Permission, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM permissions ORDER BY name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []*domain.Permission
	for rows.Next() {
		p := &domain.Permission{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	if perms == nil {
		perms = []*domain.Permission{}
	}
	return perms, rows.Err()
}

func (r *PermissionRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM permissions`).Scan(&total)
	return total, err
}

func (r *PermissionRepo) FindByID(ctx context.Context, id string) (*domain.Permission, error) {
	p := &domain.Permission{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM permissions WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("permission not found: %w", domain.ErrNotFound)
	}
	return p, err
}
