package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/features/maintenancerequest/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type MaintenanceRequestRepo struct {
	db *database.DB
}

func NewMaintenanceRequestRepo(db *database.DB) *MaintenanceRequestRepo {
	return &MaintenanceRequestRepo{db: db}
}

const maintenanceCols = `
	m.id, m.room_id, m.title, m.description, m.status, m.priority,
	m.reported_date, m.resolved_date, m.note, m.created_at, m.updated_at,
	r.room_number`

func scanDetail(row pgx.Row) (*domain.MaintenanceRequestDetail, error) {
	d := &domain.MaintenanceRequestDetail{}
	err := row.Scan(
		&d.ID, &d.RoomID, &d.Title, &d.Description, &d.Status, &d.Priority,
		&d.ReportedDate, &d.ResolvedDate, &d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.RoomNumber,
	)
	return d, err
}

func (r *MaintenanceRequestRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.MaintenanceRequest, error) {
	m := &domain.MaintenanceRequest{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT m.id, m.room_id, m.title, m.description, m.status, m.priority,
		       m.reported_date, m.resolved_date, m.note, m.created_at, m.updated_at
		FROM maintenance_requests m
		JOIN rooms r ON r.id = m.room_id
		WHERE m.id = $1 AND r.dormitory_id = $2`, id, dormitoryID).
		Scan(&m.ID, &m.RoomID, &m.Title, &m.Description, &m.Status, &m.Priority,
			&m.ReportedDate, &m.ResolvedDate, &m.Note, &m.CreatedAt, &m.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("maintenance request not found: %w", coredomain.ErrNotFound)
	}
	return m, err
}

func (r *MaintenanceRequestRepo) FindDetailByID(ctx context.Context, dormitoryID, id string) (*domain.MaintenanceRequestDetail, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT `+maintenanceCols+`
		FROM maintenance_requests m
		JOIN rooms r ON r.id = m.room_id
		WHERE m.id = $1 AND r.dormitory_id = $2`, id, dormitoryID)
	d, err := scanDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("maintenance request not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *MaintenanceRequestRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.MaintenanceRequestDetail, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT `+maintenanceCols+`
		FROM maintenance_requests m
		JOIN rooms r ON r.id = m.room_id
		WHERE r.dormitory_id = $1
		ORDER BY
			CASE m.status
				WHEN 'open'        THEN 1
				WHEN 'in_progress' THEN 2
				WHEN 'done'        THEN 3
				WHEN 'cancelled'   THEN 4
			END,
			CASE m.priority
				WHEN 'urgent' THEN 1
				WHEN 'high'   THEN 2
				WHEN 'normal' THEN 3
				WHEN 'low'    THEN 4
			END,
			m.reported_date DESC
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.MaintenanceRequestDetail
	for rows.Next() {
		d, err := scanDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.MaintenanceRequestDetail{}
	}
	return list, rows.Err()
}

func (r *MaintenanceRequestRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM maintenance_requests m
		JOIN rooms r ON r.id = m.room_id
		WHERE r.dormitory_id = $1`, dormitoryID).Scan(&total)
	return total, err
}

func (r *MaintenanceRequestRepo) Create(ctx context.Context, m *domain.MaintenanceRequest) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO maintenance_requests
			(id, room_id, title, description, status, priority, reported_date, resolved_date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		m.ID, m.RoomID, m.Title, m.Description, m.Status, m.Priority,
		m.ReportedDate, m.ResolvedDate, m.Note)
	return err
}

func (r *MaintenanceRequestRepo) Update(ctx context.Context, dormitoryID string, m *domain.MaintenanceRequest) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE maintenance_requests
		SET room_id=$3, title=$4, description=$5, status=$6, priority=$7,
		    reported_date=$8, resolved_date=$9, note=$10, updated_at=NOW()
		FROM rooms r
		WHERE maintenance_requests.id = $1 AND r.id = maintenance_requests.room_id AND r.dormitory_id = $2`,
		m.ID, dormitoryID, m.RoomID, m.Title, m.Description, m.Status, m.Priority,
		m.ReportedDate, m.ResolvedDate, m.Note)
	return err
}

func (r *MaintenanceRequestRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `
		DELETE FROM maintenance_requests m
		USING rooms r
		WHERE m.id = $1 AND r.id = m.room_id AND r.dormitory_id = $2`, id, dormitoryID)
	return err
}
