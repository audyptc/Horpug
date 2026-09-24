package postgres

import (
	"context"
	"errors"

	portaldomain "apihorpug/internal/features/tenantportal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository serves the tenant self-service pages. Every query is keyed by
// the signed-in tenant's id, so a tenant only ever reaches their own data.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindTenantByLineUserID returns the active tenant linked to a LINE account.
func (r *Repository) FindTenantByLineUserID(ctx context.Context, lineUserID string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		SELECT id FROM tenants WHERE line_user_id = $1 AND line_user_id <> '' AND is_active = TRUE
	`, lineUserID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, portaldomain.ErrNotLinked
	}
	return id, err
}

func (r *Repository) Profile(ctx context.Context, tenantID uuid.UUID) (portaldomain.Profile, error) {
	p := portaldomain.Profile{TenantID: tenantID}
	err := r.db.QueryRow(ctx, `
		SELECT first_name, last_name FROM tenants WHERE id = $1 AND is_active = TRUE
	`, tenantID).Scan(&p.FirstName, &p.LastName)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, portaldomain.ErrNotLinked
	}
	if err != nil {
		return p, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT c.id, rm.id, rm.room_number, d.id, d.name, COALESCE(d.phone, ''), c.rent_price::float8, c.start_date
		FROM contracts c
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE c.tenant_id = $1 AND c.status = 'active'
		ORDER BY d.name, rm.room_number
	`, tenantID)
	if err != nil {
		return p, err
	}
	p.Rooms, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (portaldomain.Room, error) {
		var room portaldomain.Room
		err := row.Scan(&room.ContractID, &room.RoomID, &room.RoomNumber, &room.DormitoryID, &room.DormitoryName,
			&room.DormitoryPhone, &room.RentPrice, &room.StartDate)
		return room, err
	})
	return p, err
}

// Invoices lists the tenant's invoices across all their contracts (a former
// room may still be owed on), newest period first, cancelled ones left out.
func (r *Repository) Invoices(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.Invoice, error) {
	rows, err := r.db.Query(ctx, `
		SELECT i.id, rm.room_number, d.name, i.period_year, i.period_month, i.due_date, i.total_amount::float8,
			GREATEST(i.total_amount - COALESCE((SELECT SUM(p.total_amount) FROM payments p WHERE p.invoice_id = i.id AND p.status = 'active'), 0), 0)::float8,
			i.status
		FROM invoices i
		JOIN contracts c ON c.id = i.contract_id
		JOIN rooms rm ON rm.id = c.room_id
		JOIN dormitories d ON d.id = rm.dormitory_id
		WHERE c.tenant_id = $1 AND i.status <> 'cancelled'
		ORDER BY i.period_year DESC, i.period_month DESC, rm.room_number
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (portaldomain.Invoice, error) {
		var inv portaldomain.Invoice
		err := row.Scan(&inv.ID, &inv.RoomNumber, &inv.DormitoryName, &inv.PeriodYear, &inv.PeriodMonth, &inv.DueDate,
			&inv.TotalAmount, &inv.Outstanding, &inv.Status)
		if inv.Status == "paid" {
			inv.Outstanding = 0
		}
		return inv, err
	})
}

// RepairRequests lists repairs the tenant reported, newest first.
func (r *Repository) RepairRequests(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.RepairRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rr.id, rm.room_number, rr.category, rr.description, rr.status, rr.reported_date, rr.updated_at
		FROM repair_requests rr
		JOIN rooms rm ON rm.id = rr.room_id
		WHERE rr.tenant_id = $1
		ORDER BY rr.reported_date DESC, rr.created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[portaldomain.RepairRequest])
}

// CreateRepairRequest files a repair for a room the tenant currently rents.
// It has no staff creator; tenant_id records who reported it.
func (r *Repository) CreateRepairRequest(ctx context.Context, tenantID, roomID uuid.UUID, category, description string) (portaldomain.RepairRequest, error) {
	var rents bool
	if err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM contracts WHERE tenant_id = $1 AND room_id = $2 AND status = 'active')
	`, tenantID, roomID).Scan(&rents); err != nil {
		return portaldomain.RepairRequest{}, err
	}
	if !rents {
		return portaldomain.RepairRequest{}, portaldomain.ErrRoomNotRented
	}

	id := uuid.New()
	if _, err := r.db.Exec(ctx, `
		INSERT INTO repair_requests (id, room_id, tenant_id, category, description, status, reported_date)
		VALUES ($1, $2, $3, $4, $5, 'pending', CURRENT_DATE)
	`, id, roomID, tenantID, category, description); err != nil {
		return portaldomain.RepairRequest{}, err
	}
	return r.repairRequest(ctx, id, tenantID)
}

// CancelRepairRequest withdraws a tenant's own request while staff haven't
// started on it.
func (r *Repository) CancelRepairRequest(ctx context.Context, tenantID, id uuid.UUID) (portaldomain.RepairRequest, error) {
	var status string
	err := r.db.QueryRow(ctx, `SELECT status FROM repair_requests WHERE id = $1 AND tenant_id = $2`, id, tenantID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return portaldomain.RepairRequest{}, portaldomain.ErrRepairNotFound
	}
	if err != nil {
		return portaldomain.RepairRequest{}, err
	}
	if status != "pending" {
		return portaldomain.RepairRequest{}, portaldomain.ErrRepairNotPending
	}
	if _, err := r.db.Exec(ctx, `
		UPDATE repair_requests SET status = 'cancelled', updated_at = NOW() WHERE id = $1 AND tenant_id = $2 AND status = 'pending'
	`, id, tenantID); err != nil {
		return portaldomain.RepairRequest{}, err
	}
	return r.repairRequest(ctx, id, tenantID)
}

func (r *Repository) repairRequest(ctx context.Context, id, tenantID uuid.UUID) (portaldomain.RepairRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT rr.id, rm.room_number, rr.category, rr.description, rr.status, rr.reported_date, rr.updated_at
		FROM repair_requests rr
		JOIN rooms rm ON rm.id = rr.room_id
		WHERE rr.id = $1 AND rr.tenant_id = $2
	`, id, tenantID)
	if err != nil {
		return portaldomain.RepairRequest{}, err
	}
	req, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByPos[portaldomain.RepairRequest])
	if errors.Is(err, pgx.ErrNoRows) {
		return req, portaldomain.ErrRepairNotFound
	}
	return req, err
}

// Announcements lists published announcements of the dormitories where the
// tenant currently rents, pinned first, then newest.
func (r *Repository) Announcements(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.Announcement, error) {
	rows, err := r.db.Query(ctx, `
		SELECT a.id, d.name, a.title, a.content, a.category, a.is_pinned, a.published_date
		FROM announcements a
		JOIN dormitories d ON d.id = a.dormitory_id
		WHERE a.is_published = TRUE AND a.published_date <= CURRENT_DATE
		AND a.dormitory_id IN (
			SELECT rm.dormitory_id FROM contracts c JOIN rooms rm ON rm.id = c.room_id
			WHERE c.tenant_id = $1 AND c.status = 'active'
		)
		ORDER BY a.is_pinned DESC, a.published_date DESC, a.created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[portaldomain.Announcement])
}
