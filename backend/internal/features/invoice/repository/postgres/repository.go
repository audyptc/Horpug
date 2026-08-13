package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	invoiceusecase "apihorpug/internal/features/invoice/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectInvoiceColumns = `
	i.id, i.contract_id, c.tenant_id, t.first_name, t.last_name,
	c.room_id, rm.room_number, rm.dormitory_id, d.name,
	i.period_year, i.period_month, i.issue_date, i.due_date, i.total_amount, i.status, i.paid_at, i.note,
	i.created_by, i.updated_by, i.created_at, i.updated_at
`

const invoiceFromJoins = `
	FROM invoices i
	JOIN contracts c ON c.id = i.contract_id
	JOIN tenants t ON t.id = c.tenant_id
	JOIN rooms rm ON rm.id = c.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters invoiceusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.ContractID != nil {
		conditions = append(conditions, fmt.Sprintf(`i.contract_id = $%d`, *argIdx))
		*args = append(*args, *filters.ContractID)
		*argIdx++
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf(`i.status = $%d`, *argIdx))
		*args = append(*args, *filters.Status)
		*argIdx++
	}
	if filters.PeriodYear != nil {
		conditions = append(conditions, fmt.Sprintf(`i.period_year = $%d`, *argIdx))
		*args = append(*args, *filters.PeriodYear)
		*argIdx++
	}
	if filters.PeriodMonth != nil {
		conditions = append(conditions, fmt.Sprintf(`i.period_month = $%d`, *argIdx))
		*args = append(*args, *filters.PeriodMonth)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters invoiceusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + invoiceFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters invoiceusecase.ListFilters, limit, offset int) ([]invoicedomain.Invoice, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectInvoiceColumns + invoiceFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY i.period_year DESC, i.period_month DESC, rm.room_number ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanInvoices(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Invoice, error) {
	if err := r.ensureInvoiceAccess(ctx, id, requesterID); err != nil {
		return invoicedomain.Invoice{}, err
	}

	invoice, err := r.loadInvoiceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicedomain.Invoice{}, invoicedomain.ErrInvoiceNotFound
		}
		return invoicedomain.Invoice{}, err
	}

	items, err := r.loadInvoiceItems(ctx, id)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	invoice.Items = items

	return invoice, nil
}

// Create builds an invoice for a contract's billing period by pulling the
// contract's rent price plus any electricity/water meter readings recorded
// against the contract's room within that period, and stores each as a line
// item. It runs in a transaction so the invoice and its items are consistent.
func (r *Repository) Create(ctx context.Context, input invoiceusecase.CreateInput) (invoicedomain.Invoice, error) {
	roomID, rentPrice, err := r.loadContractForInvoice(ctx, input.ContractID, input.CreatedBy)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}

	periodStart := time.Date(input.PeriodYear, time.Month(input.PeriodMonth), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	defer tx.Rollback(ctx)

	id := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO invoices (id, contract_id, period_year, period_month, issue_date, due_date, total_amount, status, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8, $9, $10)
	`, id, input.ContractID, input.PeriodYear, input.PeriodMonth, input.IssueDate, input.DueDate,
		invoicedomain.InvoiceStatusUnpaid, input.Note, input.CreatedBy, input.CreatedBy); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return invoicedomain.Invoice{}, invoicedomain.ErrInvoiceExists
			}
			if pgErr.Code == "23503" {
				return invoicedomain.Invoice{}, invoicedomain.ErrContractNotFound
			}
		}
		return invoicedomain.Invoice{}, err
	}

	total := 0.0

	rentAmount, err := insertInvoiceItem(ctx, tx, id, invoicedomain.InvoiceItemTypeRent, "ค่าเช่าห้อง", nil, rentPrice)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	total += rentAmount

	electricityTotal, err := r.appendMeterItems(ctx, tx, id, roomID, "electricity_meters", invoicedomain.InvoiceItemTypeElectricity, "ค่าไฟฟ้า", periodStart, periodEnd)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	total += electricityTotal

	waterTotal, err := r.appendMeterItems(ctx, tx, id, roomID, "water_meters", invoicedomain.InvoiceItemTypeWater, "ค่าน้ำประปา", periodStart, periodEnd)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	total += waterTotal

	if _, err := tx.Exec(ctx, `UPDATE invoices SET total_amount = $1 WHERE id = $2`, total, id); err != nil {
		return invoicedomain.Invoice{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return invoicedomain.Invoice{}, err
	}

	invoice, err := r.loadInvoiceByID(ctx, id)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	items, err := r.loadInvoiceItems(ctx, id)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	invoice.Items = items

	return invoice, nil
}

func insertInvoiceItem(ctx context.Context, tx pgx.Tx, invoiceID uuid.UUID, itemType invoicedomain.InvoiceItemType, description string, referenceID *uuid.UUID, amount float64) (float64, error) {
	_, err := tx.Exec(ctx, `
		INSERT INTO invoice_items (id, invoice_id, item_type, description, reference_id, amount)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), invoiceID, itemType, description, referenceID, amount)
	if err != nil {
		return 0, err
	}
	return amount, nil
}

// appendMeterItems inserts one invoice item per meter reading of the given
// table (electricity_meters or water_meters) recorded for roomID within
// [periodStart, periodEnd), and returns their summed amount.
func (r *Repository) appendMeterItems(ctx context.Context, tx pgx.Tx, invoiceID, roomID uuid.UUID, table string, itemType invoicedomain.InvoiceItemType, description string, periodStart, periodEnd time.Time) (float64, error) {
	rows, err := tx.Query(ctx, fmt.Sprintf(`
		SELECT id, total_amount FROM %s
		WHERE room_id = $1 AND reading_date >= $2 AND reading_date < $3
		ORDER BY reading_date ASC
	`, table), roomID, periodStart, periodEnd)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type meterRow struct {
		id     uuid.UUID
		amount float64
	}
	meters := make([]meterRow, 0)
	for rows.Next() {
		var m meterRow
		if err := rows.Scan(&m.id, &m.amount); err != nil {
			return 0, err
		}
		meters = append(meters, m)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	total := 0.0
	for _, m := range meters {
		refID := m.id
		if _, err := insertInvoiceItem(ctx, tx, invoiceID, itemType, description, &refID, m.amount); err != nil {
			return 0, err
		}
		total += m.amount
	}

	return total, nil
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input invoiceusecase.UpdateInput) (invoicedomain.Invoice, error) {
	if err := r.ensureInvoiceAccess(ctx, id, requesterID); err != nil {
		return invoicedomain.Invoice{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
		if *input.Status == invoicedomain.InvoiceStatusPaid {
			setClauses = append(setClauses, "paid_at = NOW()")
		} else {
			setClauses = append(setClauses, "paid_at = NULL")
		}
	}
	if input.DueDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("due_date = $%d", argIdx))
		args = append(args, *input.DueDate)
		argIdx++
	}
	if input.Note != nil {
		setClauses = append(setClauses, fmt.Sprintf("note = $%d", argIdx))
		args = append(args, *input.Note)
		argIdx++
	}
	if input.UpdatedBy != nil {
		setClauses = append(setClauses, fmt.Sprintf("updated_by = $%d", argIdx))
		args = append(args, *input.UpdatedBy)
		argIdx++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = NOW()")
		args = append(args, id)
		query := fmt.Sprintf("UPDATE invoices SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23514" {
				return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceStatus
			}
			return invoicedomain.Invoice{}, err
		}
	}

	invoice, err := r.loadInvoiceByID(ctx, id)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	items, err := r.loadInvoiceItems(ctx, id)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}
	invoice.Items = items

	return invoice, nil
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureInvoiceAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return invoicedomain.ErrInvoiceNotFound
	}

	return nil
}

func (r *Repository) loadInvoiceByID(ctx context.Context, id uuid.UUID) (invoicedomain.Invoice, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectInvoiceColumns+invoiceFromJoins+` WHERE i.id = $1`, id)
	return scanInvoice(row)
}

func (r *Repository) loadInvoiceItems(ctx context.Context, invoiceID uuid.UUID) ([]invoicedomain.InvoiceItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, invoice_id, item_type, description, reference_id, amount, created_at
		FROM invoice_items WHERE invoice_id = $1 ORDER BY created_at ASC
	`, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]invoicedomain.InvoiceItem, 0)
	for rows.Next() {
		var item invoicedomain.InvoiceItem
		if err := rows.Scan(&item.ID, &item.InvoiceID, &item.ItemType, &item.Description, &item.ReferenceID, &item.Amount, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanInvoice(row pgx.Row) (invoicedomain.Invoice, error) {
	var invoice invoicedomain.Invoice
	var firstName, lastName string
	if err := row.Scan(
		&invoice.ID,
		&invoice.ContractID,
		&invoice.TenantID,
		&firstName,
		&lastName,
		&invoice.RoomID,
		&invoice.RoomNumber,
		&invoice.DormitoryID,
		&invoice.DormitoryName,
		&invoice.PeriodYear,
		&invoice.PeriodMonth,
		&invoice.IssueDate,
		&invoice.DueDate,
		&invoice.TotalAmount,
		&invoice.Status,
		&invoice.PaidAt,
		&invoice.Note,
		&invoice.CreatedBy,
		&invoice.UpdatedBy,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	); err != nil {
		return invoicedomain.Invoice{}, err
	}
	invoice.TenantName = strings.TrimSpace(firstName + " " + lastName)
	return invoice, nil
}

func scanInvoices(rows pgx.Rows) ([]invoicedomain.Invoice, error) {
	invoices := make([]invoicedomain.Invoice, 0)
	for rows.Next() {
		invoice, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return invoices, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages invoices in every dormitory), along with their
// role ID so callers can also check role-level dormitory grants.
func (r *Repository) dormitoryScope(ctx context.Context, userID uuid.UUID) (full bool, roleID uuid.UUID, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&full, &roleID)
	if err != nil {
		return false, uuid.Nil, err
	}
	return full, roleID, nil
}

// loadContractForInvoice confirms the contract exists and the requester may
// bill against it, based on access to the dormitory of its room, then
// returns the room ID and rent price to build the invoice from. A nil
// requesterID (should not normally happen, since Create always supplies one)
// skips the dormitory-access check.
func (r *Repository) loadContractForInvoice(ctx context.Context, contractID uuid.UUID, requesterID *uuid.UUID) (roomID uuid.UUID, rentPrice float64, err error) {
	if requesterID == nil {
		err = r.db.QueryRow(ctx, `SELECT room_id, rent_price FROM contracts WHERE id = $1`, contractID).Scan(&roomID, &rentPrice)
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, 0, invoicedomain.ErrContractNotFound
		}
		return roomID, rentPrice, err
	}

	full, roleID, err := r.dormitoryScope(ctx, *requesterID)
	if err != nil {
		return uuid.Nil, 0, err
	}

	err = r.db.QueryRow(ctx, `
		SELECT c.room_id, c.rent_price
		FROM contracts c
		JOIN rooms rm ON rm.id = c.room_id
		WHERE c.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, contractID, full, *requesterID, roleID).Scan(&roomID, &rentPrice)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, 0, invoicedomain.ErrContractNotFound
		}
		return uuid.Nil, 0, err
	}
	return roomID, rentPrice, nil
}

// ensureInvoiceAccess confirms the invoice exists and the requester may act
// on it, based on access to the dormitory of its contract's room. Both a
// missing invoice and a missing grant surface as ErrInvoiceNotFound so
// scoped-out callers can't distinguish the two.
func (r *Repository) ensureInvoiceAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		WHERE i.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return invoicedomain.ErrInvoiceNotFound
		}
		return err
	}
	return nil
}
