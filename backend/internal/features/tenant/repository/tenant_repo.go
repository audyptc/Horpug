package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/features/tenant/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type TenantRepo struct {
	db *database.DB
}

func NewTenantRepo(db *database.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Tenant, error) {
	t := &domain.Tenant{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT t.id, t.dormitory_id, t.first_name, t.last_name, t.phone, t.id_card, t.email, t.emergency_contact, t.note,
		       ARRAY(
		           SELECT r.room_number
		           FROM contracts c
		           JOIN rooms r ON r.id = c.room_id
		           WHERE c.tenant_id = t.id AND c.status = 'active' AND c.deleted_at IS NULL
		           ORDER BY r.room_number
		       ),
		       COALESCE(t.created_by::text, ''), COALESCE(t.updated_by::text, ''),
		       COALESCE(u.full_name, ''),
		       t.created_at, t.updated_at
		FROM tenants t
		LEFT JOIN users u ON u.id = t.updated_by
		WHERE t.id = $1 AND t.dormitory_id = $2 AND t.deleted_at IS NULL`, id, dormitoryID).
		Scan(&t.ID, &t.DormitoryID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note,
			&t.ActiveRoomNumbers,
			&t.CreatedBy, &t.UpdatedBy, &t.UpdatedByName,
			&t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %w", coredomain.ErrNotFound)
	}
	return t, err
}

func (r *TenantRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT t.id, t.dormitory_id, t.first_name, t.last_name, t.phone, t.id_card, t.email, t.emergency_contact, t.note,
		       ARRAY(
		           SELECT r.room_number
		           FROM contracts c
		           JOIN rooms r ON r.id = c.room_id
		           WHERE c.tenant_id = t.id AND c.status = 'active' AND c.deleted_at IS NULL
		           ORDER BY r.room_number
		       ),
		       COALESCE(t.created_by::text, ''), COALESCE(t.updated_by::text, ''),
		       COALESCE(u.full_name, ''),
		       t.created_at, t.updated_at
		FROM tenants t
		LEFT JOIN users u ON u.id = t.updated_by
		WHERE t.dormitory_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.first_name, t.last_name
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t := &domain.Tenant{}
		if err := rows.Scan(&t.ID, &t.DormitoryID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note,
			&t.ActiveRoomNumbers,
			&t.CreatedBy, &t.UpdatedBy, &t.UpdatedByName,
			&t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, t)
	}
	if tenants == nil {
		tenants = []*domain.Tenant{}
	}
	return tenants, rows.Err()
}

func (r *TenantRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tenants WHERE dormitory_id = $1 AND deleted_at IS NULL`, dormitoryID).Scan(&total)
	return total, err
}

func (r *TenantRepo) Create(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO tenants (id, dormitory_id, first_name, last_name, phone, id_card, email, emergency_contact, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, '')::uuid, NULLIF($10, '')::uuid)`,
		t.ID, t.DormitoryID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note, t.CreatedBy)
	return err
}

func (r *TenantRepo) Update(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE tenants
		SET first_name = $3, last_name = $4, phone = $5, id_card = $6,
		    email = $7, emergency_contact = $8, note = $9,
		    updated_by = NULLIF($10, '')::uuid, updated_at = NOW()
		WHERE id = $1 AND dormitory_id = $2`,
		t.ID, t.DormitoryID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note, t.UpdatedBy)
	return err
}

func (r *TenantRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tenants SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND dormitory_id = $2 AND deleted_at IS NULL`,
		id, dormitoryID)
	return err
}
