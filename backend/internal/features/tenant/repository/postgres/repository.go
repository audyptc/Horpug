package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tenantdomain "apihorpug/internal/features/tenant/domain"
	tenantusecase "apihorpug/internal/features/tenant/usecase"

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

func (r *Repository) Count(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]tenantdomain.Tenant, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, is_active, created_by, updated_by, created_at, updated_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTenants(rows)
}

func (r *Repository) ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error) {
	query := `
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, is_active, created_by, updated_by, created_at, updated_at
		FROM tenants
		WHERE is_active = true
	`
	args := make([]any, 0)
	argIdx := 1
	if search != "" {
		query += fmt.Sprintf(` AND (first_name ILIKE $%d OR last_name ILIKE $%d OR phone ILIKE $%d OR id_card ILIKE $%d)`, argIdx, argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	query += fmt.Sprintf(` ORDER BY first_name ASC, last_name ASC LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTenants(rows)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error) {
	tenant, err := r.loadTenantByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tenantdomain.Tenant{}, tenantdomain.ErrTenantNotFound
		}
		return tenantdomain.Tenant{}, err
	}
	return tenant, nil
}

func (r *Repository) Create(ctx context.Context, input tenantusecase.CreateInput) (tenantdomain.Tenant, error) {
	tenant := tenantdomain.Tenant{
		ID:               uuid.New(),
		FirstName:        input.FirstName,
		LastName:         input.LastName,
		Phone:            input.Phone,
		IDCard:           input.IDCard,
		Email:            input.Email,
		EmergencyContact: input.EmergencyContact,
		Note:             input.Note,
		IsActive:         input.IsActive,
		CreatedBy:        input.CreatedBy,
		UpdatedBy:        input.CreatedBy,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO tenants (id, first_name, last_name, phone, id_card, email, emergency_contact, note, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`, tenant.ID, tenant.FirstName, tenant.LastName, tenant.Phone, tenant.IDCard, tenant.Email, tenant.EmergencyContact, tenant.Note, tenant.IsActive, tenant.CreatedBy, tenant.UpdatedBy).
		Scan(&tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return tenantdomain.Tenant{}, tenantdomain.ErrTenantIDCardExists
		}
		return tenantdomain.Tenant{}, err
	}

	return tenant, nil
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input tenantusecase.UpdateInput) (tenantdomain.Tenant, error) {
	if err := r.ensureTenantExists(ctx, id); err != nil {
		return tenantdomain.Tenant{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argIdx))
		args = append(args, *input.FirstName)
		argIdx++
	}
	if input.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argIdx))
		args = append(args, *input.LastName)
		argIdx++
	}
	if input.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *input.Phone)
		argIdx++
	}
	if input.IDCard != nil {
		setClauses = append(setClauses, fmt.Sprintf("id_card = $%d", argIdx))
		args = append(args, *input.IDCard)
		argIdx++
	}
	if input.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *input.Email)
		argIdx++
	}
	if input.EmergencyContact != nil {
		setClauses = append(setClauses, fmt.Sprintf("emergency_contact = $%d", argIdx))
		args = append(args, *input.EmergencyContact)
		argIdx++
	}
	if input.Note != nil {
		setClauses = append(setClauses, fmt.Sprintf("note = $%d", argIdx))
		args = append(args, *input.Note)
		argIdx++
	}
	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *input.IsActive)
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
		query := fmt.Sprintf("UPDATE tenants SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return tenantdomain.Tenant{}, tenantdomain.ErrTenantIDCardExists
			}
			return tenantdomain.Tenant{}, err
		}
	}

	return r.loadTenantByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM tenants WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return tenantdomain.ErrTenantNotFound
	}

	return nil
}

func (r *Repository) ensureTenantExists(ctx context.Context, id uuid.UUID) error {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM tenants WHERE id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return tenantdomain.ErrTenantNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) loadTenantByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error) {
	var tenant tenantdomain.Tenant
	err := r.db.QueryRow(ctx, `
		SELECT id, first_name, last_name, phone, id_card, email, emergency_contact, note, is_active, created_by, updated_by, created_at, updated_at
		FROM tenants
		WHERE id = $1
	`, id).Scan(
		&tenant.ID,
		&tenant.FirstName,
		&tenant.LastName,
		&tenant.Phone,
		&tenant.IDCard,
		&tenant.Email,
		&tenant.EmergencyContact,
		&tenant.Note,
		&tenant.IsActive,
		&tenant.CreatedBy,
		&tenant.UpdatedBy,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		return tenantdomain.Tenant{}, err
	}
	return tenant, nil
}

func scanTenants(rows pgx.Rows) ([]tenantdomain.Tenant, error) {
	tenants := make([]tenantdomain.Tenant, 0)
	for rows.Next() {
		var tenant tenantdomain.Tenant
		if err := rows.Scan(
			&tenant.ID,
			&tenant.FirstName,
			&tenant.LastName,
			&tenant.Phone,
			&tenant.IDCard,
			&tenant.Email,
			&tenant.EmergencyContact,
			&tenant.Note,
			&tenant.IsActive,
			&tenant.CreatedBy,
			&tenant.UpdatedBy,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tenants, nil
}
