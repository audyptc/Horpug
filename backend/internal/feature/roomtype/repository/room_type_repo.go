package repository

import (
	"context"
	"fmt"

	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/roomtype/domain"
	"apigofiberhorpug/internal/platform/database"

	"github.com/jackc/pgx/v5"
)

type RoomTypeRepo struct {
	db *database.DB
}

func NewRoomTypeRepo(db *database.DB) *RoomTypeRepo {
	return &RoomTypeRepo{db: db}
}

func (r *RoomTypeRepo) List(ctx context.Context) ([]*domain.RoomType, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, sort_order, created_at, updated_at
		FROM room_types
		ORDER BY sort_order, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var types []*domain.RoomType
	for rows.Next() {
		rt := &domain.RoomType{}
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.SortOrder, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		types = append(types, rt)
	}
	if types == nil {
		types = []*domain.RoomType{}
	}
	return types, rows.Err()
}

func (r *RoomTypeRepo) FindByID(ctx context.Context, id string) (*domain.RoomType, error) {
	rt := &domain.RoomType{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, sort_order, created_at, updated_at
		FROM room_types WHERE id = $1`, id).
		Scan(&rt.ID, &rt.Name, &rt.SortOrder, &rt.CreatedAt, &rt.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("room type not found: %w", coredomain.ErrNotFound)
	}
	return rt, err
}

func (r *RoomTypeRepo) Create(ctx context.Context, rt *domain.RoomType) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO room_types (id, name, sort_order)
		VALUES ($1, $2, $3)`,
		rt.ID, rt.Name, rt.SortOrder)
	return err
}

func (r *RoomTypeRepo) Update(ctx context.Context, rt *domain.RoomType) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE room_types
		SET name = $2, sort_order = $3, updated_at = NOW()
		WHERE id = $1`,
		rt.ID, rt.Name, rt.SortOrder)
	return err
}

func (r *RoomTypeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM room_types WHERE id = $1`, id)
	return err
}
