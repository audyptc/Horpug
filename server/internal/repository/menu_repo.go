package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type MenuRepo struct {
	db *database.DB
}

func NewMenuRepo(db *database.DB) *MenuRepo {
	return &MenuRepo{db: db}
}

func (r *MenuRepo) FindByID(ctx context.Context, id string) (*domain.Menu, error) {
	m := &domain.Menu{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, path, created_at, updated_at
		FROM menus WHERE id = $1`, id).
		Scan(&m.ID, &m.Name, &m.Path, &m.CreatedAt, &m.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("menu not found")
	}
	return m, err
}

func (r *MenuRepo) List(ctx context.Context) ([]*domain.Menu, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, path, created_at, updated_at
		FROM menus ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []*domain.Menu
	for rows.Next() {
		m := &domain.Menu{}
		if err := rows.Scan(&m.ID, &m.Name, &m.Path, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		menus = append(menus, m)
	}
	if menus == nil {
		menus = []*domain.Menu{}
	}
	return menus, nil
}
