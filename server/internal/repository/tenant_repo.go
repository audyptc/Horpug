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
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, created_at, updated_at
		FROM tenants WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&t.ID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note, &t.CreatedAt, &t.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %w", domain.ErrNotFound)
	}
	return t, err
}

func (r *TenantRepo) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, created_at, updated_at
		FROM tenants
		WHERE deleted_at IS NULL
		ORDER BY first_name, last_name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		t := &domain.Tenant{}
		if err := rows.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Phone, &t.IDCard,
			&t.Email, &t.EmergencyContact, &t.Note, &t.CreatedAt, &t.UpdatedAt); err != nil {
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
		INSERT INTO tenants (id, first_name, last_name, phone, id_card, email, emergency_contact, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.ID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note)
	return err
}

func (r *TenantRepo) Update(ctx context.Context, t *domain.Tenant) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE tenants
		SET first_name = $2, last_name = $3, phone = $4, id_card = $5,
		    email = $6, emergency_contact = $7, note = $8, updated_at = NOW()
		WHERE id = $1`,
		t.ID, t.FirstName, t.LastName, t.Phone, t.IDCard, t.Email, t.EmergencyContact, t.Note)
	return err
}

func (r *TenantRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE tenants SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
