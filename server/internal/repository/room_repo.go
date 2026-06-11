package repository

import (
	"context"
	"errors"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
		SELECT r.id, r.room_number, r.floor, r.type, r.status, r.rent_price, r.description,
		       COALESCE(r.created_by::text, ''), COALESCE(r.updated_by::text, ''),
		       COALESCE(u.full_name, ''),
		       r.created_at, r.updated_at
		FROM rooms r
		LEFT JOIN users u ON u.id = r.updated_by
		WHERE r.id = $1 AND r.deleted_at IS NULL`, id).
		Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Type, &room.Status,
			&room.RentPrice, &room.Description, &room.CreatedBy, &room.UpdatedBy,
			&room.UpdatedByName, &room.CreatedAt, &room.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("room not found: %w", domain.ErrNotFound)
	}
	return room, err
}

func (r *RoomRepo) List(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT r.id, r.room_number, r.floor, r.type, r.status, r.rent_price, r.description,
		       COALESCE(r.created_by::text, ''), COALESCE(r.updated_by::text, ''),
		       COALESCE(u.full_name, ''),
		       r.created_at, r.updated_at
		FROM rooms r
		LEFT JOIN users u ON u.id = r.updated_by
		WHERE r.deleted_at IS NULL
		ORDER BY r.floor, r.room_number
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*domain.Room
	for rows.Next() {
		room := &domain.Room{}
		if err := rows.Scan(&room.ID, &room.RoomNumber, &room.Floor, &room.Type, &room.Status,
			&room.RentPrice, &room.Description, &room.CreatedBy, &room.UpdatedBy,
			&room.UpdatedByName, &room.CreatedAt, &room.UpdatedAt); err != nil {
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
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM rooms WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *RoomRepo) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO rooms (id, room_number, floor, type, status, rent_price, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, '')::uuid, NULLIF($8, '')::uuid)`,
		room.ID, room.RoomNumber, room.Floor, room.Type, room.Status, room.RentPrice, room.Description, room.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("room number already exists: %w", domain.ErrDuplicate)
		}
	}
	return err
}

func (r *RoomRepo) Update(ctx context.Context, room *domain.Room) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE rooms
		SET room_number = $2, floor = $3, type = $4, status = $5, rent_price = $6, description = $7,
		    updated_by = NULLIF($8, '')::uuid, updated_at = NOW()
		WHERE id = $1`,
		room.ID, room.RoomNumber, room.Floor, room.Type, room.Status, room.RentPrice, room.Description, room.UpdatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return fmt.Errorf("room number already exists: %w", domain.ErrDuplicate)
		}
	}
	return err
}

func (r *RoomRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE rooms SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
