package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/feature/bill/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

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
		b.rent_amount, b.electric_amount, b.water_amount, b.parking_amount, b.other_amount,
		b.total_amount, b.status, b.due_date, b.paid_at, b.note,
		b.created_at, b.updated_at,
		COALESCE(b.created_by::text, ''), COALESCE(b.updated_by::text, ''), COALESCE(ub.full_name, ''),
		t.first_name, t.last_name, r.room_number
	FROM bills b
	JOIN contracts c ON c.id = b.contract_id
	JOIN tenants   t ON t.id = c.tenant_id
	JOIN rooms     r ON r.id = c.room_id
	LEFT JOIN users ub ON ub.id = b.updated_by`

func scanBillDetail(row pgx.Row) (*domain.BillDetail, error) {
	d := &domain.BillDetail{}
	err := row.Scan(
		&d.ID, &d.ContractID, &d.BillingMonth,
		&d.RentAmount, &d.ElectricAmount, &d.WaterAmount, &d.ParkingAmount, &d.OtherAmount,
		&d.TotalAmount, &d.Status, &d.DueDate, &d.PaidAt, &d.Note,
		&d.CreatedAt, &d.UpdatedAt,
		&d.CreatedBy, &d.UpdatedBy, &d.UpdatedByName,
		&d.TenantFirstName, &d.TenantLastName, &d.RoomNumber,
	)
	d.OtherItems = []domain.BillOtherItem{}
	return d, err
}

func (r *BillRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Bill, error) {
	b := &domain.Bill{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT b.id, b.contract_id, b.billing_month,
		       b.rent_amount, b.electric_amount, b.water_amount, b.parking_amount, b.other_amount,
		       b.total_amount, b.status, b.due_date, b.paid_at, b.note, b.created_at, b.updated_at,
		       COALESCE(b.created_by::text, ''), COALESCE(b.updated_by::text, '')
		FROM bills b
		JOIN contracts c ON c.id = b.contract_id
		JOIN rooms r ON r.id = c.room_id
		WHERE b.id = $1 AND r.dormitory_id = $2 AND b.deleted_at IS NULL`, id, dormitoryID).
		Scan(&b.ID, &b.ContractID, &b.BillingMonth,
			&b.RentAmount, &b.ElectricAmount, &b.WaterAmount, &b.ParkingAmount, &b.OtherAmount,
			&b.TotalAmount, &b.Status, &b.DueDate, &b.PaidAt, &b.Note,
			&b.CreatedAt, &b.UpdatedAt,
			&b.CreatedBy, &b.UpdatedBy)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bill not found: %w", coredomain.ErrNotFound)
	}
	return b, err
}

func (r *BillRepo) FindDetailByID(ctx context.Context, dormitoryID, id string) (*domain.BillDetail, error) {
	row := r.db.Pool.QueryRow(ctx, billDetailSelect+` WHERE b.id = $1 AND r.dormitory_id = $2 AND b.deleted_at IS NULL`, id, dormitoryID)
	d, err := scanBillDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("bill not found: %w", coredomain.ErrNotFound)
	}
	if err != nil {
		return nil, err
	}
	items, err := r.FindOtherItems(ctx, id)
	if err != nil {
		return nil, err
	}
	d.OtherItems = items
	return d, nil
}

func (r *BillRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.BillDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		billDetailSelect+` WHERE r.dormitory_id = $1 AND b.deleted_at IS NULL ORDER BY b.billing_month DESC, b.created_at DESC LIMIT $2 OFFSET $3`,
		dormitoryID, limit, offset)
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
		return []*domain.BillDetail{}, nil
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]string, len(list))
	for i, b := range list {
		ids[i] = b.ID
	}
	itemRows, err := r.db.Pool.Query(ctx,
		`SELECT id, bill_id, label, amount, sort_order FROM bill_other_items WHERE bill_id = ANY($1) ORDER BY sort_order`,
		ids)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	idxMap := make(map[string]int, len(list))
	for i, b := range list {
		idxMap[b.ID] = i
	}
	for itemRows.Next() {
		var item domain.BillOtherItem
		if err := itemRows.Scan(&item.ID, &item.BillID, &item.Label, &item.Amount, &item.SortOrder); err != nil {
			return nil, err
		}
		idx := idxMap[item.BillID]
		list[idx].OtherItems = append(list[idx].OtherItems, item)
	}
	return list, itemRows.Err()
}

func (r *BillRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM bills b
		JOIN contracts c ON c.id = b.contract_id
		JOIN rooms r ON r.id = c.room_id
		WHERE r.dormitory_id = $1 AND b.deleted_at IS NULL`, dormitoryID).Scan(&total)
	return total, err
}

func (r *BillRepo) Create(ctx context.Context, b *domain.Bill) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO bills (
			id, contract_id, billing_month,
			rent_amount, electric_amount, water_amount, parking_amount, other_amount,
			total_amount, status, due_date, note, created_by, updated_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, NULLIF($13,'')::uuid, NULLIF($13,'')::uuid)`,
		b.ID, b.ContractID, b.BillingMonth,
		b.RentAmount, b.ElectricAmount, b.WaterAmount, b.ParkingAmount, b.OtherAmount,
		b.TotalAmount, b.Status, b.DueDate, b.Note, b.CreatedBy)
	return err
}

func (r *BillRepo) Update(ctx context.Context, dormitoryID string, b *domain.Bill) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE bills
		SET rent_amount=$3, electric_amount=$4, water_amount=$5, parking_amount=$6,
		    other_amount=$7, total_amount=$8,
		    status=$9, due_date=$10, paid_at=$11, note=$12,
		    updated_by=NULLIF($13,'')::uuid, updated_at=NOW()
		FROM contracts c, rooms r
		WHERE bills.id = $1 AND c.id = bills.contract_id AND r.id = c.room_id AND r.dormitory_id = $2`,
		b.ID, dormitoryID, b.RentAmount, b.ElectricAmount, b.WaterAmount, b.ParkingAmount,
		b.OtherAmount, b.TotalAmount,
		b.Status, b.DueDate, b.PaidAt, b.Note, b.UpdatedBy)
	return err
}

func (r *BillRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE bills SET deleted_at = NOW(), updated_at = NOW()
		FROM contracts c, rooms r
		WHERE bills.id = $1 AND c.id = bills.contract_id AND r.id = c.room_id
		  AND r.dormitory_id = $2 AND bills.deleted_at IS NULL`, id, dormitoryID)
	return err
}

func (r *BillRepo) ReplaceOtherItems(ctx context.Context, billID string, items []domain.BillOtherItem) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM bill_other_items WHERE bill_id = $1`, billID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO bill_other_items (id, bill_id, label, amount, sort_order)
			VALUES ($1, $2, $3, $4, $5)`,
			item.ID, billID, item.Label, item.Amount, item.SortOrder); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *BillRepo) FindOtherItems(ctx context.Context, billID string) ([]domain.BillOtherItem, error) {
	rows, err := r.db.Pool.Query(ctx,
		`SELECT id, bill_id, label, amount, sort_order FROM bill_other_items WHERE bill_id = $1 ORDER BY sort_order`,
		billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.BillOtherItem
	for rows.Next() {
		var item domain.BillOtherItem
		if err := rows.Scan(&item.ID, &item.BillID, &item.Label, &item.Amount, &item.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []domain.BillOtherItem{}
	}
	return items, rows.Err()
}
