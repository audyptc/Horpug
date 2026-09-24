package postgres

import (
	"context"
	"strings"
	"time"

	reportdomain "apihorpug/internal/features/report/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// Query is a report filter resolved into bind parameters. Every report query
// starts with the four scope parameters —
//
//	$1 full dormitory access  $2 requester id  $3 role id  $4 dormitory filter (NULL = all)
//
// — and appends only what else it uses (Postgres rejects parameters a query
// doesn't reference), via dates(), period() or all().
type Query struct {
	full        bool
	requesterID uuid.UUID
	roleID      uuid.UUID
	dormitoryID *uuid.UUID
	start, end  time.Time
	year, month int
}

func (q Query) scopeArgs() []any { return []any{q.full, q.requesterID, q.roleID, q.dormitoryID} }

// dates binds $5 month start and $6 next month start.
func (q Query) dates() []any { return append(q.scopeArgs(), q.start, q.end) }

// period binds $5 period year and $6 period month.
func (q Query) period() []any { return append(q.scopeArgs(), q.year, q.month) }

// all binds $5/$6 as dates() and $7/$8 as the period.
func (q Query) all() []any { return append(q.dates(), q.year, q.month) }

// scope restricts rows to dormitories the requester manages (and to the
// requested one, if any).
func scope(column string) string {
	return strings.ReplaceAll(`($1 OR {col} IN (
		SELECT dormitory_id FROM user_dormitories WHERE user_id = $2
		UNION
		SELECT dormitory_id FROM role_dormitories WHERE role_id = $3
	)) AND ($4::uuid IS NULL OR {col} = $4)`, "{col}", column)
}

func (r *Repository) newQuery(ctx context.Context, f reportdomain.Filter) (Query, error) {
	q := Query{requesterID: f.RequesterID, dormitoryID: f.DormitoryID, year: f.Year, month: f.Month}
	if err := r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = $1
	`, f.RequesterID).Scan(&q.full, &q.roleID); err != nil {
		return Query{}, err
	}
	q.start = time.Date(f.Year, time.Month(f.Month), 1, 0, 0, 0, 0, time.UTC)
	q.end = q.start.AddDate(0, 1, 0)
	return q, nil
}

// monthPayments are active payments dated in [$5, $6).
const monthPayments = `
	SELECT p.id, p.dormitory_id, p.total_amount::float8 AS amount
	FROM payments p
	WHERE p.status = 'active' AND p.payment_date >= $5 AND p.payment_date < $6`

// monthExpenses are expenses dated in [$5, $6).
const monthExpenses = `
	SELECT e.dormitory_id, e.category, e.amount::float8 AS amount
	FROM expenses e
	WHERE e.expense_date >= $5 AND e.expense_date < $6`

// periodInvoices are the invoices of the billing period (year and month
// parameters given), cancelled ones excluded, with what active payments have
// collected on each.
func periodInvoices(yearParam, monthParam string) string {
	return `
	SELECT i.id, rm.dormitory_id, i.total_amount::float8 AS total,
		COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0)::float8 AS paid
	FROM invoices i
	JOIN contracts c ON c.id = i.contract_id
	JOIN rooms rm ON rm.id = c.room_id
	WHERE i.period_year = ` + yearParam + ` AND i.period_month = ` + monthParam + ` AND i.status <> 'cancelled'`
}

func (r *Repository) amounts(ctx context.Context, sql string, args []any) ([]reportdomain.Amount, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (reportdomain.Amount, error) {
		var a reportdomain.Amount
		err := row.Scan(&a.Key, &a.Amount)
		return a, err
	})
}

// Monthly fills the report's figures.
func (r *Repository) Monthly(ctx context.Context, f reportdomain.Filter) (reportdomain.MonthlyReport, error) {
	q, err := r.newQuery(ctx, f)
	if err != nil {
		return reportdomain.MonthlyReport{}, err
	}
	rep := reportdomain.MonthlyReport{Year: q.year, Month: q.month}

	if err = r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM (`+monthPayments+`) mp WHERE `+scope("mp.dormitory_id"),
		q.dates()...).Scan(&rep.Income.Count, &rep.Income.Total); err != nil {
		return rep, err
	}
	if rep.Income.ByMethod, err = r.amounts(ctx, `
		SELECT pi.payment_method, SUM(pi.amount)::float8
		FROM (`+monthPayments+`) mp
		JOIN payment_items pi ON pi.payment_id = mp.id
		WHERE `+scope("mp.dormitory_id")+`
		GROUP BY pi.payment_method ORDER BY 2 DESC`, q.dates()); err != nil {
		return rep, err
	}

	invoices := periodInvoices("$5", "$6")
	if err = r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total), 0), COALESCE(SUM(LEAST(paid, total)), 0), COALESCE(SUM(GREATEST(total - paid, 0)), 0)
		FROM (`+invoices+`) pi WHERE `+scope("pi.dormitory_id"),
		q.period()...).Scan(&rep.Billing.InvoiceCount, &rep.Billing.Billed, &rep.Billing.Collected, &rep.Billing.Outstanding); err != nil {
		return rep, err
	}
	if rep.Billing.ByItemType, err = r.amounts(ctx, `
		SELECT ii.item_type, SUM(ii.amount)::float8
		FROM (`+invoices+`) pi
		JOIN invoice_items ii ON ii.invoice_id = pi.id
		WHERE `+scope("pi.dormitory_id")+`
		GROUP BY ii.item_type ORDER BY 2 DESC`, q.period()); err != nil {
		return rep, err
	}

	if err = r.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(amount), 0) FROM (`+monthExpenses+`) me WHERE `+scope("me.dormitory_id"),
		q.dates()...).Scan(&rep.Expenses.Total); err != nil {
		return rep, err
	}
	if rep.Expenses.ByCategory, err = r.amounts(ctx, `
		SELECT me.category, SUM(me.amount)::float8 FROM (`+monthExpenses+`) me
		WHERE `+scope("me.dormitory_id")+`
		GROUP BY me.category ORDER BY 2 DESC`, q.dates()); err != nil {
		return rep, err
	}

	if err = r.db.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(owed), 0) FROM (
			SELECT rm.dormitory_id,
				(i.total_amount - COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0))::float8 AS owed
			FROM invoices i
			JOIN contracts c ON c.id = i.contract_id
			JOIN rooms rm ON rm.id = c.room_id
			WHERE i.status IN ('unpaid', 'overdue')
		) a WHERE a.owed > 0.005 AND `+scope("a.dormitory_id"),
		q.scopeArgs()...).Scan(&rep.Arrears.Count, &rep.Arrears.Amount); err != nil {
		return rep, err
	}

	if err = r.db.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE rm.status = 'occupied')
		FROM rooms rm WHERE rm.is_active AND `+scope("rm.dormitory_id"),
		q.scopeArgs()...).Scan(&rep.Occupancy.Rooms, &rep.Occupancy.Occupied); err != nil {
		return rep, err
	}

	breakdown := periodInvoices("$7", "$8")
	rows, err := r.db.Query(ctx, `
		SELECT d.id, d.name,
			COALESCE((SELECT SUM(amount) FROM (`+monthPayments+`) mp WHERE mp.dormitory_id = d.id), 0),
			COALESCE((SELECT SUM(amount) FROM (`+monthExpenses+`) me WHERE me.dormitory_id = d.id), 0),
			COALESCE((SELECT SUM(total) FROM (`+breakdown+`) pi WHERE pi.dormitory_id = d.id), 0),
			COALESCE((SELECT SUM(GREATEST(total - paid, 0)) FROM (`+breakdown+`) pi WHERE pi.dormitory_id = d.id), 0)
		FROM dormitories d
		WHERE `+scope("d.id")+`
		ORDER BY d.name`, q.all()...)
	if err != nil {
		return rep, err
	}
	rep.Dormitories, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (reportdomain.DormitoryRow, error) {
		var d reportdomain.DormitoryRow
		err := row.Scan(&d.ID, &d.Name, &d.Income, &d.Expenses, &d.Billed, &d.Outstanding)
		return d, err
	})
	return rep, err
}

// Details lists the rows behind the report for export.
func (r *Repository) Details(ctx context.Context, f reportdomain.Filter) (reportdomain.Details, error) {
	var d reportdomain.Details
	q, err := r.newQuery(ctx, f)
	if err != nil {
		return d, err
	}

	// Voided payments are listed too, marked, so the receipt sequence has no
	// unexplained gaps; they aren't counted in the totals.
	payments, err := r.db.Query(ctx, `
		SELECT COALESCE(p.receipt_no, ''), p.payment_date, dm.name, rm.room_number, t.first_name || ' ' || t.last_name,
			i.period_year, i.period_month,
			COALESCE((SELECT string_agg(DISTINCT pi.payment_method, ', ') FROM payment_items pi WHERE pi.payment_id = p.id), ''),
			p.total_amount::float8, p.status = 'voided'
		FROM payments p
		JOIN invoices i ON i.id = p.invoice_id
		JOIN contracts c ON c.id = i.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories dm ON dm.id = rm.dormitory_id
		WHERE p.payment_date >= $5 AND p.payment_date < $6 AND `+scope("rm.dormitory_id")+`
		ORDER BY dm.name, p.receipt_no`, q.dates()...)
	if err != nil {
		return d, err
	}
	d.Payments, err = pgx.CollectRows(payments, func(row pgx.CollectableRow) (reportdomain.PaymentLine, error) {
		var l reportdomain.PaymentLine
		err := row.Scan(&l.ReceiptNo, &l.PaymentDate, &l.DormitoryName, &l.RoomNumber, &l.TenantName,
			&l.PeriodYear, &l.PeriodMonth, &l.Methods, &l.Amount, &l.Voided)
		return l, err
	})
	if err != nil {
		return d, err
	}

	invoices, err := r.db.Query(ctx, `
		SELECT dm.name, rm.room_number, t.first_name || ' ' || t.last_name, i.due_date,
			i.total_amount::float8,
			COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0)::float8,
			i.status
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN tenants t ON t.id = c.tenant_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories dm ON dm.id = rm.dormitory_id
		WHERE i.period_year = $5 AND i.period_month = $6 AND `+scope("rm.dormitory_id")+`
		ORDER BY dm.name, rm.room_number`, q.period()...)
	if err != nil {
		return d, err
	}
	d.Invoices, err = pgx.CollectRows(invoices, func(row pgx.CollectableRow) (reportdomain.InvoiceLine, error) {
		var l reportdomain.InvoiceLine
		err := row.Scan(&l.DormitoryName, &l.RoomNumber, &l.TenantName, &l.DueDate, &l.Total, &l.Paid, &l.Status)
		return l, err
	})
	if err != nil {
		return d, err
	}

	expenses, err := r.db.Query(ctx, `
		SELECT e.expense_date, dm.name, e.category, COALESCE(e.description, ''), e.amount::float8
		FROM expenses e
		JOIN dormitories dm ON dm.id = e.dormitory_id
		WHERE e.expense_date >= $5 AND e.expense_date < $6 AND `+scope("e.dormitory_id")+`
		ORDER BY e.expense_date, dm.name`, q.dates()...)
	if err != nil {
		return d, err
	}
	d.Expenses, err = pgx.CollectRows(expenses, func(row pgx.CollectableRow) (reportdomain.ExpenseLine, error) {
		var l reportdomain.ExpenseLine
		err := row.Scan(&l.ExpenseDate, &l.DormitoryName, &l.Category, &l.Description, &l.Amount)
		return l, err
	})
	return d, err
}
