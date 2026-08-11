package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/features/contract/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type ContractRepo struct {
	db *database.DB
}

func NewContractRepo(db *database.DB) *ContractRepo {
	return &ContractRepo{db: db}
}

const contractDetailSelect = `
	SELECT
		c.id, c.tenant_id, c.room_id, c.start_date, c.end_date,
		c.rent_price, c.deposit, c.num_occupants, c.status, c.note, c.created_at, c.updated_at,
		COALESCE(c.created_by::text, ''), COALESCE(c.updated_by::text, ''),
		t.first_name, t.last_name, r.room_number,
		COALESCE(u.full_name, '')
	FROM contracts c
	JOIN tenants t ON t.id = c.tenant_id
	JOIN rooms   r ON r.id = c.room_id
	LEFT JOIN users u ON u.id = c.updated_by`

func scanContractDetail(row pgx.Row) (*domain.ContractDetail, error) {
	d := &domain.ContractDetail{}
	err := row.Scan(
		&d.ID, &d.TenantID, &d.RoomID, &d.StartDate, &d.EndDate,
		&d.RentPrice, &d.Deposit, &d.NumOccupants, &d.Status, &d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.CreatedBy, &d.UpdatedBy,
		&d.TenantFirstName, &d.TenantLastName, &d.RoomNumber,
		&d.UpdatedByName,
	)
	return d, err
}

func (r *ContractRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Contract, error) {
	c := &domain.Contract{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT c.id, c.tenant_id, c.room_id, c.start_date, c.end_date,
		       c.rent_price, c.deposit, c.num_occupants, c.status, c.note, c.created_at, c.updated_at,
		       COALESCE(c.created_by::text, ''), COALESCE(c.updated_by::text, '')
		FROM contracts c
		JOIN rooms r ON r.id = c.room_id
		WHERE c.id = $1 AND r.dormitory_id = $2 AND c.deleted_at IS NULL`, id, dormitoryID).
		Scan(&c.ID, &c.TenantID, &c.RoomID, &c.StartDate, &c.EndDate,
			&c.RentPrice, &c.Deposit, &c.NumOccupants, &c.Status, &c.Note, &c.CreatedAt, &c.UpdatedAt,
			&c.CreatedBy, &c.UpdatedBy)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("contract not found: %w", coredomain.ErrNotFound)
	}
	return c, err
}

func (r *ContractRepo) FindDetailByID(ctx context.Context, dormitoryID, id string) (*domain.ContractDetail, error) {
	row := r.db.Pool.QueryRow(ctx,
		contractDetailSelect+` WHERE c.id = $1 AND r.dormitory_id = $2 AND c.deleted_at IS NULL`, id, dormitoryID)
	d, err := scanContractDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("contract not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *ContractRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.ContractDetail, error) {
	rows, err := r.db.Pool.Query(ctx,
		contractDetailSelect+` WHERE r.dormitory_id = $1 AND c.deleted_at IS NULL ORDER BY c.created_at DESC LIMIT $2 OFFSET $3`,
		dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ContractDetail
	for rows.Next() {
		d, err := scanContractDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.ContractDetail{}
	}
	return list, rows.Err()
}

func (r *ContractRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM contracts c
		JOIN rooms r ON r.id = c.room_id
		WHERE r.dormitory_id = $1 AND c.deleted_at IS NULL`, dormitoryID).Scan(&total)
	return total, err
}

func (r *ContractRepo) Create(ctx context.Context, c *domain.Contract) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO contracts (id, tenant_id, room_id, start_date, end_date, rent_price, deposit, num_occupants, status, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NULLIF($11, '')::uuid, NULLIF($11, '')::uuid)`,
		c.ID, c.TenantID, c.RoomID, c.StartDate, c.EndDate, c.RentPrice, c.Deposit, c.NumOccupants, c.Status, c.Note, c.CreatedBy)
	return err
}

func (r *ContractRepo) Update(ctx context.Context, dormitoryID string, c *domain.Contract) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE contracts SET end_date = $3, rent_price = $4, deposit = $5, num_occupants = $6, status = $7, note = $8,
		    updated_by = NULLIF($9, '')::uuid, updated_at = NOW()
		FROM rooms r
		WHERE contracts.id = $1 AND r.id = contracts.room_id AND r.dormitory_id = $2`,
		c.ID, dormitoryID, c.EndDate, c.RentPrice, c.Deposit, c.NumOccupants, c.Status, c.Note, c.UpdatedBy)
	return err
}

func (r *ContractRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE contracts SET deleted_at = NOW(), updated_at = NOW()
		FROM rooms r
		WHERE contracts.id = $1 AND r.id = contracts.room_id AND r.dormitory_id = $2 AND contracts.deleted_at IS NULL`,
		id, dormitoryID)
	return err
}

func (r *ContractRepo) HasActiveContractForRoom(ctx context.Context, roomID string) (bool, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM contracts
		WHERE room_id = $1 AND status = 'active' AND deleted_at IS NULL`, roomID).Scan(&count)
	return count > 0, err
}

func (r *ContractRepo) HasActiveContractForTenant(ctx context.Context, tenantID string) (bool, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM contracts
		WHERE tenant_id = $1 AND status = 'active' AND deleted_at IS NULL`, tenantID).Scan(&count)
	return count > 0, err
}
