package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/feature/parcel/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

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
		&p.ID, &p.DormitoryID, &p.TrackingNumber, &p.RecipientName, &p.RoomNumber,
		&p.Status, &p.ReceivedDate, &p.PickedUpDate,
		&p.Note, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

const selectParcel = `
	SELECT
		id, dormitory_id, tracking_number, recipient_name, room_number,
		status, received_date, picked_up_date,
		note, created_at, updated_at
	FROM parcels`

func (r *ParcelRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Parcel, error) {
	row := r.db.Pool.QueryRow(ctx, selectParcel+` WHERE id = $1 AND dormitory_id = $2`, id, dormitoryID)
	p, err := scanParcel(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("parcel not found: %w", coredomain.ErrNotFound)
	}
	return p, err
}

func (r *ParcelRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Parcel, error) {
	rows, err := r.db.Pool.Query(ctx, selectParcel+`
		WHERE dormitory_id = $1
		ORDER BY received_date DESC, created_at DESC
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
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

func (r *ParcelRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM parcels WHERE dormitory_id = $1`, dormitoryID).Scan(&total)
	return total, err
}

func (r *ParcelRepo) Create(ctx context.Context, p *domain.Parcel) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO parcels (id, dormitory_id, tracking_number, recipient_name, room_number, status, received_date, picked_up_date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.DormitoryID, p.TrackingNumber, p.RecipientName, p.RoomNumber,
		p.Status, p.ReceivedDate, p.PickedUpDate, p.Note)
	return err
}

func (r *ParcelRepo) Update(ctx context.Context, p *domain.Parcel) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE parcels
		SET tracking_number=$3, recipient_name=$4, room_number=$5, status=$6,
		    received_date=$7, picked_up_date=$8, note=$9, updated_at=NOW()
		WHERE id=$1 AND dormitory_id=$2`,
		p.ID, p.DormitoryID, p.TrackingNumber, p.RecipientName, p.RoomNumber,
		p.Status, p.ReceivedDate, p.PickedUpDate, p.Note)
	return err
}

func (r *ParcelRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM parcels WHERE id = $1 AND dormitory_id = $2`, id, dormitoryID)
	return err
}
