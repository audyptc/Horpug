package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/parcel/domain"

	"github.com/jackc/pgx/v5"
)

type ParcelRepo struct {
	db *database.DB
}

func NewParcelRepo(db *database.DB) *ParcelRepo {
	return &ParcelRepo{db: db}
}

func scanParcel(row pgx.Row) (*domain.Parcel, error) {
	p := &domain.Parcel{}
	err := row.Scan(
		&p.ID, &p.TrackingNumber, &p.RecipientName, &p.RoomNumber,
		&p.Status, &p.ReceivedDate, &p.PickedUpDate,
		&p.Note, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

const selectParcel = `
	SELECT
		id, tracking_number, recipient_name, room_number,
		status, received_date, picked_up_date,
		note, created_at, updated_at
	FROM parcels`

func (r *ParcelRepo) FindByID(ctx context.Context, id string) (*domain.Parcel, error) {
	row := r.db.Pool.QueryRow(ctx, selectParcel+` WHERE id = $1`, id)
	p, err := scanParcel(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("parcel not found: %w", coredomain.ErrNotFound)
	}
	return p, err
}

func (r *ParcelRepo) List(ctx context.Context, limit, offset int) ([]*domain.Parcel, error) {
	rows, err := r.db.Pool.Query(ctx, selectParcel+`
		ORDER BY received_date DESC, created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Parcel
	for rows.Next() {
		p, err := scanParcel(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	if list == nil {
		list = []*domain.Parcel{}
	}
	return list, rows.Err()
}

func (r *ParcelRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM parcels`).Scan(&total)
	return total, err
}

func (r *ParcelRepo) Create(ctx context.Context, p *domain.Parcel) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO parcels (id, tracking_number, recipient_name, room_number, status, received_date, picked_up_date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		p.ID, p.TrackingNumber, p.RecipientName, p.RoomNumber,
		p.Status, p.ReceivedDate, p.PickedUpDate, p.Note)
	return err
}

func (r *ParcelRepo) Update(ctx context.Context, p *domain.Parcel) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE parcels
		SET tracking_number=$2, recipient_name=$3, room_number=$4, status=$5,
		    received_date=$6, picked_up_date=$7, note=$8, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.TrackingNumber, p.RecipientName, p.RoomNumber,
		p.Status, p.ReceivedDate, p.PickedUpDate, p.Note)
	return err
}

func (r *ParcelRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM parcels WHERE id = $1`, id)
	return err
}
