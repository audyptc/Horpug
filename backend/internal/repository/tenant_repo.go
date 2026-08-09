package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

	"github.com/jackc/pgx/v5"
)

type TenantRepo struct {
	db *database.DB
}

func NewTenantRepo(db *database.DB) *TenantRepo {
	return &TenantRepo{db: db}
}

func (r *TenantRepo) FindByID(ctx context.Context, id string) (*domain.Tenant, error) {
	t := &domain.Tenant{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT t.id, t.first_name, t.last_name, t.phone, t.id_card, t.email, t.emergency_contact, t.note,
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
		WHERE t.id = $1 AND t.deleted_at IS NULL`, id).
		Scan(&t.ID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note,
			&t.ActiveRoomNumbers,
			&t.CreatedBy, &t.UpdatedBy, &t.UpdatedByName,
			&t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %w", domain.ErrNotFound)
	}
	return t, err
}

func (r *TenantRepo) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT t.id, t.first_name, t.last_name, t.phone, t.id_card, t.email, t.emergency_contact, t.note,
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
		WHERE t.deleted_at IS NULL
		ORDER BY t.first_name, t.last_name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t := &domain.Tenant{}
		if err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
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

func (r *TenantRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM tenants WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *TenantRepo) Create(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO tenants (id, first_name, last_name, phone, id_card, email, emergency_contact, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULLIF($9, '')::uuid, NULLIF($9, '')::uuid)`,
		t.ID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note, t.CreatedBy)
	return err
}

func (r *TenantRepo) Update(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE tenants
		SET first_name = $2, last_name = $3, phone = $4, id_card = $5,
		    email = $6, emergency_contact = $7, note = $8,
		    updated_by = NULLIF($9, '')::uuid, updated_at = NOW()
		WHERE id = $1`,
		t.ID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note, t.UpdatedBy)
	return err
}

func (r *TenantRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tenants SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
