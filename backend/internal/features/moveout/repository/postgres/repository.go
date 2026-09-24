package postgres

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	moveoutdomain "apihorpug/internal/features/moveout/domain"
	moveoutusecase "apihorpug/internal/features/moveout/usecase"
	paymentrepository "apihorpug/internal/features/payment/repository/postgres"

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

// querier is what both the pool and a transaction offer, so the settlement
// inputs can be read the same way for a preview and inside the confirming
// transaction.
type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// dormitoryAccess is the SQL predicate for "the requester may act on
// dormitory $n": full-access roles, or a direct or role-level grant.
const dormitoryAccess = `($2 OR EXISTS (
	SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = %s AND ud.user_id = $3
) OR EXISTS (
	SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = %s AND rd.role_id = $4
))`

func (r *Repository) dormitoryScope(ctx context.Context, q querier, userID uuid.UUID) (full bool, roleID uuid.UUID, err error) {
	err = q.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&full, &roleID)
	return full, roleID, err
}

// loadBase reads the contract a settlement is for, checking the requester
// manages its dormitory; FOR UPDATE when confirming, so two move-outs of the
// same contract can't run at once. A contract outside the requester's reach
// surfaces as ErrContractNotFound.
func (r *Repository) loadBase(ctx context.Context, q querier, contractID, requesterID uuid.UUID, lock bool) (moveoutdomain.Settlement, string, error) {
	full, roleID, err := r.dormitoryScope(ctx, q, requesterID)
	if err != nil {
		return moveoutdomain.Settlement{}, "", err
	}

	query := `
		SELECT c.id, t.first_name || ' ' || t.last_name, rm.id, rm.room_number, d.id, d.name,
			c.start_date, c.rent_price::float8, c.deposit::float8, c.status
		FROM contracts c
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE c.id = $1 AND ` + sprintfAccess("rm.dormitory_id")
	if lock {
		query += ` FOR UPDATE OF c`
	}

	var s moveoutdomain.Settlement
	var status string
	err = q.QueryRow(ctx, query, contractID, full, requesterID, roleID).Scan(
		&s.ContractID, &s.TenantName, &s.RoomID, &s.RoomNumber, &s.DormitoryID, &s.DormitoryName,
		&s.StartDate, &s.RentPrice, &s.Deposit, &status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return moveoutdomain.Settlement{}, "", moveoutdomain.ErrContractNotFound
		}
		return moveoutdomain.Settlement{}, "", err
	}
	return s, status, nil
}

// sprintfAccess fills the dormitory column into dormitoryAccess.
func sprintfAccess(column string) string {
	return strings.ReplaceAll(dormitoryAccess, "%s", column)
}

// loadInputs gathers what the settlement is built from: the contract's unpaid
// invoice balances (active payments only), the rent already billed for the
// move-out month, and meter readings in the contract up to the move-out date
// that nothing has billed yet.
func (r *Repository) loadInputs(ctx context.Context, q querier, base moveoutdomain.Settlement, moveOutDate time.Time) (moveoutusecase.Inputs, error) {
	in := moveoutusecase.Inputs{
		OutstandingInvoices: make([]moveoutusecase.OutstandingInvoice, 0),
		UnbilledReadings:    make([]moveoutusecase.Reading, 0),
	}

	rows, err := q.Query(ctx, `
		SELECT i.id, i.period_year, i.period_month,
			(i.total_amount - COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0))::float8
		FROM invoices i
		WHERE i.contract_id = $1 AND i.status IN ('unpaid', 'overdue')
		ORDER BY i.period_year, i.period_month, i.id
	`, base.ContractID)
	if err != nil {
		return in, err
	}
	for rows.Next() {
		var inv moveoutusecase.OutstandingInvoice
		if err := rows.Scan(&inv.ID, &inv.PeriodYear, &inv.PeriodMonth, &inv.Outstanding); err != nil {
			rows.Close()
			return in, err
		}
		in.OutstandingInvoices = append(in.OutstandingInvoices, inv)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return in, err
	}

	var monthInvoiceExists bool
	var rentBilled float64
	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) > 0, COALESCE(SUM(ii.amount) FILTER (WHERE ii.item_type = 'rent'), 0)::float8
		FROM invoices i
		LEFT JOIN invoice_items ii ON ii.invoice_id = i.id
		WHERE i.contract_id = $1 AND i.period_year = $2 AND i.period_month = $3 AND i.status <> 'cancelled'
	`, base.ContractID, moveOutDate.Year(), int(moveOutDate.Month())).Scan(&monthInvoiceExists, &rentBilled); err != nil {
		return in, err
	}
	if monthInvoiceExists {
		in.MoveOutMonthRentBilled = &rentBilled
	}

	for _, meter := range []struct {
		table string
		kind  moveoutdomain.ItemType
	}{
		{"electricity_meters", moveoutdomain.ItemElectricity},
		{"water_meters", moveoutdomain.ItemWater},
	} {
		rows, err := q.Query(ctx, `
			SELECT m.id, m.reading_date, m.total_amount::float8
			FROM `+meter.table+` m
			WHERE m.room_id = $1 AND m.reading_date >= $2 AND m.reading_date <= $3
			AND NOT EXISTS (SELECT 1 FROM invoice_items ii WHERE ii.reference_id = m.id)
			AND NOT EXISTS (SELECT 1 FROM move_out_items mi WHERE mi.reference_id = m.id)
			ORDER BY m.reading_date
		`, base.RoomID, base.StartDate, moveOutDate)
		if err != nil {
			return in, err
		}
		for rows.Next() {
			reading := moveoutusecase.Reading{Kind: meter.kind}
			if err := rows.Scan(&reading.ID, &reading.ReadingDate, &reading.Amount); err != nil {
				rows.Close()
				return in, err
			}
			in.UnbilledReadings = append(in.UnbilledReadings, reading)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return in, err
		}
	}

	return in, nil
}

// LoadForPreview reads the settlement inputs without changing anything.
func (r *Repository) LoadForPreview(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time) (moveoutdomain.Settlement, moveoutusecase.Inputs, error) {
	base, status, err := r.loadBase(ctx, r.db, contractID, requesterID, false)
	if err != nil {
		return moveoutdomain.Settlement{}, moveoutusecase.Inputs{}, err
	}
	if status != "active" {
		return moveoutdomain.Settlement{}, moveoutusecase.Inputs{}, moveoutdomain.ErrContractNotActive
	}
	in, err := r.loadInputs(ctx, r.db, base, moveOutDate)
	return base, in, err
}

// Settle confirms a move-out in one transaction: it re-reads the inputs under
// lock, builds the settlement with build, records it, pays the contract's
// unpaid invoices from the deposit (oldest first, each a "deposit" payment
// with its own receipt number), terminates the contract and frees the room.
func (r *Repository) Settle(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time, note string,
	build func(base moveoutdomain.Settlement, in moveoutusecase.Inputs) (moveoutdomain.Settlement, error),
) (uuid.UUID, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	base, status, err := r.loadBase(ctx, tx, contractID, requesterID, true)
	if err != nil {
		return uuid.Nil, err
	}
	var alreadyMovedOut bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM move_outs WHERE contract_id = $1)`, contractID).Scan(&alreadyMovedOut); err != nil {
		return uuid.Nil, err
	}
	if alreadyMovedOut {
		return uuid.Nil, moveoutdomain.ErrAlreadyMovedOut
	}
	if status != "active" {
		return uuid.Nil, moveoutdomain.ErrContractNotActive
	}

	// Hold the contract's invoices so no payment lands between reading the
	// balances and settling them.
	if _, err := tx.Exec(ctx, `SELECT 1 FROM invoices WHERE contract_id = $1 FOR UPDATE`, contractID); err != nil {
		return uuid.Nil, err
	}

	in, err := r.loadInputs(ctx, tx, base, moveOutDate)
	if err != nil {
		return uuid.Nil, err
	}
	s, err := build(base, in)
	if err != nil {
		return uuid.Nil, err
	}

	moveOutID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO move_outs (id, contract_id, dormitory_id, move_out_date, deposit, rent_price, days_stayed, days_in_month,
			total_deductions, refund_amount, amount_due, note, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, moveOutID, contractID, s.DormitoryID, s.MoveOutDate, s.Deposit, s.RentPrice, s.DaysStayed, s.DaysInMonth,
		s.TotalDeductions, s.RefundAmount, s.AmountDue, note, requesterID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, moveoutdomain.ErrAlreadyMovedOut
		}
		return uuid.Nil, err
	}

	remaining := s.Deposit
	for i, item := range s.Items {
		var paymentID *uuid.UUID
		if item.ItemType == moveoutdomain.ItemInvoice && item.ReferenceID != nil {
			apply := math.Round(math.Min(item.Amount, remaining)*100) / 100
			if apply >= 0.01 {
				id, err := r.payFromDeposit(ctx, tx, *item.ReferenceID, s.DormitoryID, s.MoveOutDate, apply, item.Amount, requesterID)
				if err != nil {
					return uuid.Nil, err
				}
				paymentID = &id
				remaining = math.Round((remaining-apply)*100) / 100
			}
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO move_out_items (id, move_out_id, sort_order, item_type, description, amount, reference_id, payment_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, uuid.New(), moveOutID, i, item.ItemType, item.Description, item.Amount, item.ReferenceID, paymentID); err != nil {
			return uuid.Nil, err
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE contracts SET status = 'terminated', end_date = $2, updated_by = $3, updated_at = NOW() WHERE id = $1
	`, contractID, s.MoveOutDate, requesterID); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE rooms SET status = 'available', updated_at = NOW()
		WHERE id = $1 AND status = 'occupied'
		AND NOT EXISTS (SELECT 1 FROM contracts WHERE room_id = $1 AND status = 'active')
	`, s.RoomID); err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return moveOutID, nil
}

// payFromDeposit records amount against an invoice as a "deposit" payment
// with its own receipt number, marking the invoice paid when it clears the
// outstanding balance.
func (r *Repository) payFromDeposit(ctx context.Context, tx pgx.Tx, invoiceID, dormitoryID uuid.UUID, date time.Time, amount, outstanding float64, requesterID uuid.UUID) (uuid.UUID, error) {
	year := date.Year()
	seq, receiptNo, err := paymentrepository.AllocateReceipt(ctx, tx, dormitoryID, year)
	if err != nil {
		return uuid.Nil, err
	}

	paymentID := uuid.New()
	if _, err := tx.Exec(ctx, `
		INSERT INTO payments (id, invoice_id, payment_date, total_amount, note, created_by,
			dormitory_id, receipt_year, receipt_seq, receipt_no)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, paymentID, invoiceID, date, amount, "หักจากเงินประกันเมื่อย้ายออก", requesterID,
		dormitoryID, year, seq, receiptNo); err != nil {
		return uuid.Nil, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO payment_items (id, payment_id, payment_method, amount, reference_no)
		VALUES ($1, $2, 'deposit', $3, '')
	`, uuid.New(), paymentID, amount); err != nil {
		return uuid.Nil, err
	}
	if amount >= outstanding-0.005 {
		if _, err := tx.Exec(ctx, `UPDATE invoices SET status = 'paid', paid_at = NOW(), updated_at = NOW() WHERE id = $1`, invoiceID); err != nil {
			return uuid.Nil, err
		}
	}
	return paymentID, nil
}

// GetByID loads a confirmed move-out for display or printing.
func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (moveoutdomain.MoveOut, error) {
	full, roleID, err := r.dormitoryScope(ctx, r.db, requesterID)
	if err != nil {
		return moveoutdomain.MoveOut{}, err
	}

	var m moveoutdomain.MoveOut
	err = r.db.QueryRow(ctx, `
		SELECT mo.id, c.id, t.first_name || ' ' || t.last_name, rm.id, rm.room_number, d.id, d.name,
			COALESCE(d.address, ''), COALESCE(d.phone, ''),
			c.start_date, mo.move_out_date, mo.rent_price::float8, mo.deposit::float8, mo.days_stayed, mo.days_in_month,
			mo.total_deductions::float8, mo.refund_amount::float8, mo.amount_due::float8, mo.note, mo.created_by, mo.created_at
		FROM move_outs mo
		JOIN contracts c ON c.id = mo.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE mo.id = $1 AND `+sprintfAccess("rm.dormitory_id"), id, full, requesterID, roleID).Scan(
		&m.ID, &m.ContractID, &m.TenantName, &m.RoomID, &m.RoomNumber, &m.DormitoryID, &m.DormitoryName,
		&m.DormitoryAddress, &m.DormitoryPhone,
		&m.StartDate, &m.MoveOutDate, &m.RentPrice, &m.Deposit, &m.DaysStayed, &m.DaysInMonth,
		&m.TotalDeductions, &m.RefundAmount, &m.AmountDue, &m.Note, &m.CreatedBy, &m.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return moveoutdomain.MoveOut{}, moveoutdomain.ErrMoveOutNotFound
		}
		return moveoutdomain.MoveOut{}, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT mi.item_type, mi.description, mi.amount::float8, mi.reference_id,
			p.id, COALESCE(p.receipt_no, ''), COALESCE(p.total_amount, 0)::float8
		FROM move_out_items mi
		LEFT JOIN payments p ON p.id = mi.payment_id
		WHERE mi.move_out_id = $1
		ORDER BY mi.sort_order
	`, id)
	if err != nil {
		return moveoutdomain.MoveOut{}, err
	}
	defer rows.Close()

	m.Items = make([]moveoutdomain.Item, 0)
	m.DepositPayments = make([]moveoutdomain.DepositPayment, 0)
	for rows.Next() {
		var item moveoutdomain.Item
		var paymentID *uuid.UUID
		var receiptNo string
		var paid float64
		if err := rows.Scan(&item.ItemType, &item.Description, &item.Amount, &item.ReferenceID, &paymentID, &receiptNo, &paid); err != nil {
			return moveoutdomain.MoveOut{}, err
		}
		m.Items = append(m.Items, item)
		if paymentID != nil && item.ReferenceID != nil {
			m.DepositPayments = append(m.DepositPayments, moveoutdomain.DepositPayment{
				InvoiceID: *item.ReferenceID, PaymentID: *paymentID, ReceiptNo: receiptNo, Amount: paid,
			})
		}
	}
	return m, rows.Err()
}

// ListPresets returns a dormitory's ready-made deductions.
func (r *Repository) ListPresets(ctx context.Context, dormitoryID, requesterID uuid.UUID) ([]moveoutdomain.Preset, error) {
	if err := r.ensureDormitoryAccess(ctx, r.db, dormitoryID, requesterID); err != nil {
		return nil, err
	}
	return r.loadPresets(ctx, dormitoryID)
}

// PresetsForDormitory lists a dormitory's presets without an access check,
// for callers that already checked access (e.g. through the contract).
func (r *Repository) PresetsForDormitory(ctx context.Context, dormitoryID uuid.UUID) ([]moveoutdomain.Preset, error) {
	return r.loadPresets(ctx, dormitoryID)
}

func (r *Repository) loadPresets(ctx context.Context, dormitoryID uuid.UUID) ([]moveoutdomain.Preset, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, amount::float8 FROM deduction_presets WHERE dormitory_id = $1 ORDER BY sort_order, name
	`, dormitoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	presets := make([]moveoutdomain.Preset, 0)
	for rows.Next() {
		var p moveoutdomain.Preset
		if err := rows.Scan(&p.ID, &p.Name, &p.Amount); err != nil {
			return nil, err
		}
		presets = append(presets, p)
	}
	return presets, rows.Err()
}

// ReplacePresets sets a dormitory's ready-made deductions to exactly presets,
// in the given order.
func (r *Repository) ReplacePresets(ctx context.Context, dormitoryID, requesterID uuid.UUID, presets []moveoutdomain.Preset) ([]moveoutdomain.Preset, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if err := r.ensureDormitoryAccess(ctx, tx, dormitoryID, requesterID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM deduction_presets WHERE dormitory_id = $1`, dormitoryID); err != nil {
		return nil, err
	}
	for i, p := range presets {
		if _, err := tx.Exec(ctx, `
			INSERT INTO deduction_presets (id, dormitory_id, name, amount, sort_order) VALUES ($1, $2, $3, $4, $5)
		`, uuid.New(), dormitoryID, p.Name, p.Amount, i); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.loadPresets(ctx, dormitoryID)
}

func (r *Repository) ensureDormitoryAccess(ctx context.Context, q querier, dormitoryID, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, q, requesterID)
	if err != nil {
		return err
	}
	var exists int
	err = q.QueryRow(ctx, `SELECT 1 FROM dormitories d WHERE d.id = $1 AND `+sprintfAccess("d.id"),
		dormitoryID, full, requesterID, roleID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return moveoutdomain.ErrDormitoryNotFound
	}
	return err
}
