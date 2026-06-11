package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type PaymentRepo struct {
	db *database.DB
}

func NewPaymentRepo(db *database.DB) *PaymentRepo {
	return &PaymentRepo{db: db}
}

const paymentCols = `
	p.id, p.bill_id, p.amount, p.method, p.payment_date, p.note, p.created_at, p.updated_at,
	r.room_number, t.first_name, t.last_name,
	TO_CHAR(b.billing_month, 'YYYY-MM-DD'), b.total_amount`

func scanPaymentDetail(row pgx.Row) (*domain.PaymentDetail, error) {
	d := &domain.PaymentDetail{}
	err := row.Scan(
		&d.ID, &d.BillID, &d.Amount, &d.Method, &d.PaymentDate, &d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.RoomNumber, &d.TenantFirstName, &d.TenantLastName,
		&d.BillingMonth, &d.BillTotal,
	)
	return d, err
}

func (r *PaymentRepo) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	p := &domain.Payment{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, bill_id, amount, method, payment_date, note, created_at, updated_at
		FROM payments WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&p.ID, &p.BillID, &p.Amount, &p.Method, &p.PaymentDate, &p.Note, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("payment not found: %w", domain.ErrNotFound)
	}
	return p, err
}

func (r *PaymentRepo) FindDetailByID(ctx context.Context, id string) (*domain.PaymentDetail, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT `+paymentCols+`
		FROM payments p
		JOIN bills b ON b.id = p.bill_id
		JOIN contracts c ON c.id = b.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms r ON r.id = c.room_id
		WHERE p.id = $1 AND p.deleted_at IS NULL`, id)
	d, err := scanPaymentDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("payment not found: %w", domain.ErrNotFound)
	}
	return d, err
}

func (r *PaymentRepo) List(ctx context.Context, limit, offset int) ([]*domain.PaymentDetail, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT `+paymentCols+`
		FROM payments p
		JOIN bills b ON b.id = p.bill_id
		JOIN contracts c ON c.id = b.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms r ON r.id = c.room_id
		WHERE p.deleted_at IS NULL
		ORDER BY p.payment_date DESC, p.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.PaymentDetail
	for rows.Next() {
		d, err := scanPaymentDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.PaymentDetail{}
	}
	return list, rows.Err()
}

func (r *PaymentRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM payments WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *PaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (id, bill_id, amount, method, payment_date, note)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID, p.BillID, p.Amount, p.Method, p.PaymentDate, p.Note)
	if err != nil {
		return err
	}

	var billTotal float64
	err = tx.QueryRow(ctx, `SELECT total_amount FROM bills WHERE id = $1`, p.BillID).Scan(&billTotal)
	if err != nil {
		return err
	}
	if p.Amount >= billTotal {
		_, err = tx.Exec(ctx, `
			UPDATE bills SET status = 'paid', paid_at = $1, updated_at = NOW() WHERE id = $2`,
			p.PaymentDate, p.BillID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE payments
		SET amount=$2, method=$3, payment_date=$4, note=$5, updated_at=NOW()
		WHERE id=$1`,
		p.ID, p.Amount, p.Method, p.PaymentDate, p.Note)
	return err
}

func (r *PaymentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE payments SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
