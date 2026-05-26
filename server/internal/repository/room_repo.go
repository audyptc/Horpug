package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type RoomRepo struct {
	db *database.DB
}

func NewRoomRepo(db *database.DB) *RoomRepo {
	return &RoomRepo{db: db}
}

func (r *RoomRepo) FindByID(ctx context.Context, id string) (*domain.Room, error) {
	room := &domain.Room{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, room_number, floor, type, status, rent_price, description, created_at, updated_at
		FROM rooms WHERE id = $1`, id).
		Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Type, &room.Status,
			&room.RentPrice, &room.Description, &room.CreatedAt, &room.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("room not found: %w", domain.ErrNotFound)
	}
	return room, err
}

func (r *RoomRepo) List(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, room_number, floor, type, status, rent_price, description, created_at, updated_at
		FROM rooms
		ORDER BY floor, room_number
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Type, &room.Status,
			&room.RentPrice, &room.Description, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if rooms == nil {
		rooms = []*domain.Room{}
	}
	return rooms, rows.Err()
}

func (r *RoomRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rooms`).Scan(&total)
	return total, err
}

func (r *RoomRepo) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO rooms (id, room_number, floor, type, status, rent_price, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		room.ID, room.RoomNumber, room.Floor, room.Type, room.Status, room.RentPrice, room.Description)
	return err
}

func (r *RoomRepo) Update(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE rooms
		SET room_number = $2, floor = $3, type = $4, status = $5, rent_price = $6, description = $7, updated_at = NOW()
		WHERE id = $1`,
		room.ID, room.RoomNumber, room.Floor, room.Type, room.Status, room.RentPrice, room.Description)
	return err
}

func (r *RoomRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	return err
}
