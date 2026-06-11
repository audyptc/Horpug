package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type BillRepo struct {
	db *database.DB
}

func NewBillRepo(db *database.DB) *BillRepo {
	return &BillRepo{db: db}
}

const billDetailSelect = `
	SELECT
		b.id, b.contract_id, b.billing_month,
		b.rent_amount, b.electric_amount, b.water_amount, b.other_amount, b.other_note,
		b.total_amount, b.status, b.due_date, b.paid_at, b.note,
		b.created_at, b.updated_at,
		t.first_name, t.last_name, r.room_number
	FROM bills b
	JOIN contracts c ON c.id = b.contract_id
	JOIN tenants   t ON t.id = c.tenant_id
	JOIN rooms     r ON r.id = c.room_id`

func scanBillDetail(row pgx.Row) (*domain.BillDetail, error) {
	d := &domain.BillDetail{}
	err := row.Scan(
		&d.ID, &d.ContractID, &d.BillingMonth,
		&d.RentAmount, &d.ElectricAmount, &d.WaterAmount, &d.OtherAmount, &d.OtherNote,
		&d.TotalAmount, &d.Status, &d.DueDate, &d.PaidAt, &d.Note,
		&d.CreatedAt, &d.UpdatedAt,
		&d.TenantFirstName, &d.TenantLastName, &d.RoomNumber,
	)
	return d, err
}

func (r *BillRepo) FindByID(ctx context.Context, id string) (*domain.Bill, error) {
	b := &domain.Bill{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, contract_id, billing_month,
		       rent_amount, electric_amount, water_amount, other_amount, other_note,
		       total_amount, status, due_date, paid_at, note, created_at, updated_at
		FROM bills WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&b.ID, &b.ContractID, &b.BillingMonth,
			&b.RentAmount, &b.ElectricAmount, &b.WaterAmount, &b.OtherAmount, &b.OtherNote,
			&b.TotalAmount, &b.Status, &b.DueDate, &b.PaidAt, &b.Note,
			&b.CreatedAt, &b.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bill not found: %w", domain.ErrNotFound)
	}
	return b, err
}

func (r *BillRepo) FindDetailByID(ctx context.Context, id string) (*domain.BillDetail, error) {
	row := r.db.Pool.QueryRow(ctx, billDetailSelect+` WHERE b.id = $1 AND b.deleted_at IS NULL`, id)
	d, err := scanBillDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bill not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *BillRepo) List(ctx context.Context, limit, offset int) ([]*domain.BillDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		billDetailSelect+` WHERE b.deleted_at IS NULL ORDER BY b.billing_month DESC, b.created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.BillDetail
	for rows.Next() {
		d, err := scanBillDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.BillDetail{}
	}
	return list, rows.Err()
}

func (r *BillRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM bills WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *BillRepo) Create(ctx context.Context, b *domain.Bill) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO bills (
			id, contract_id, billing_month,
			rent_amount, electric_amount, water_amount, other_amount, other_note,
			total_amount, status, due_date, note
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		b.ID, b.ContractID, b.BillingMonth,
		b.RentAmount, b.ElectricAmount, b.WaterAmount, b.OtherAmount, b.OtherNote,
		b.TotalAmount, b.Status, b.DueDate, b.Note)
	return err
}

func (r *BillRepo) Update(ctx context.Context, b *domain.Bill) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE bills
		SET rent_amount=$2, electric_amount=$3, water_amount=$4,
		    other_amount=$5, other_note=$6, total_amount=$7,
		    status=$8, due_date=$9, paid_at=$10, note=$11, updated_at=NOW()
		WHERE id=$1`,
		b.ID, b.RentAmount, b.ElectricAmount, b.WaterAmount,
		b.OtherAmount, b.OtherNote, b.TotalAmount,
		b.Status, b.DueDate, b.PaidAt, b.Note)
	return err
}

func (r *BillRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE bills SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
