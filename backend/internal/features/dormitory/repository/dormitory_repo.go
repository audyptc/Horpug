package repository

import (
	"context"
	"errors"
	"fmt"

	"apigofiberhorpug/internal/features/dormitory/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DormitoryRepo struct {
	db *database.DB
}

func NewDormitoryRepo(db *database.DB) *DormitoryRepo {
	return &DormitoryRepo{db: db}
}

func (r *DormitoryRepo) FindByID(ctx context.Context, id string) (*domain.Dormitory, error) {
	d := &domain.Dormitory{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, name, address, is_active, created_at, updated_at
		FROM dormitories WHERE id = $1`, id).
		Scan(&d.ID, &d.Name, &d.Address, &d.IsActive, &d.CreatedAt, &d.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("dormitory not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *DormitoryRepo) List(ctx context.Context, limit, offset int) ([]*domain.Dormitory, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, name, address, is_active, created_at, updated_at
		FROM dormitories ORDER BY name
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dormitories []*domain.Dormitory
	for rows.Next() {
		d := &domain.Dormitory{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Address, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		dormitories = append(dormitories, d)
	}
	if dormitories == nil {
		dormitories = []*domain.Dormitory{}
	}
	return dormitories, rows.Err()
}

func (r *DormitoryRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM dormitories`).Scan(&total)
	return total, err
}

func (r *DormitoryRepo) Create(ctx context.Context, d *domain.Dormitory) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO dormitories (id, name, address, is_active) VALUES ($1, $2, $3, $4)`,
		d.ID, d.Name, d.Address, d.IsActive)
	return err
}

func (r *DormitoryRepo) Update(ctx context.Context, d *domain.Dormitory) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE dormitories SET name = $2, address = $3, is_active = $4, updated_at = NOW()
		WHERE id = $1`,
		d.ID, d.Name, d.Address, d.IsActive)
	return err
}

func (r *DormitoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM dormitories WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return fmt.Errorf("dormitory still in use: %w", coredomain.ErrDuplicate)
		}
	}
	return err
}

func (r *DormitoryRepo) ListForUser(ctx context.Context, userID string) ([]*domain.Dormitory, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT DISTINCT d.id, d.name, d.address, d.is_active, d.created_at, d.updated_at
		FROM dormitories d
		JOIN user_dormitory_roles udr ON udr.dormitory_id = d.id
		JOIN roles r ON r.id = udr.role_id
		WHERE udr.user_id = $1 AND r.is_active = TRUE
		ORDER BY d.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dormitories []*domain.Dormitory
	for rows.Next() {
		d := &domain.Dormitory{}
		if err := rows.Scan(&d.ID, &d.Name, &d.Address, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		dormitories = append(dormitories, d)
	}
	if dormitories == nil {
		dormitories = []*domain.Dormitory{}
	}
	return dormitories, rows.Err()
}

func (r *DormitoryRepo) ListAssignmentsForUser(ctx context.Context, userID string) ([]*domain.DormitoryRoleAssignment, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT d.id, d.name, r.id, r.name
		FROM user_dormitory_roles udr
		JOIN dormitories d ON d.id = udr.dormitory_id
		JOIN roles r ON r.id = udr.role_id
		WHERE udr.user_id = $1
		ORDER BY d.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*domain.DormitoryRoleAssignment
	for rows.Next() {
		item := &domain.DormitoryRoleAssignment{}
		if err := rows.Scan(&item.DormitoryID, &item.DormitoryName, &item.RoleID, &item.RoleName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []*domain.DormitoryRoleAssignment{}
	}
	return items, rows.Err()
}

func (r *DormitoryRepo) HasAccess(ctx context.Context, userID, dormitoryID string) (bool, error) {
	var exists bool
	err := r.db.Pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM user_dormitory_roles udr
			JOIN roles r ON r.id = udr.role_id
			WHERE udr.user_id = $1 AND udr.dormitory_id = $2 AND r.is_active = TRUE
		)`,
		userID, dormitoryID).Scan(&exists)
	return exists, err
}

func (r *DormitoryRepo) SetUserDormitoryRoles(ctx context.Context, userID string, items []domain.DormitoryRoleAssignmentItem) error {
	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM user_dormitory_roles WHERE user_id = $1`, userID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM user_dormitories WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_dormitory_roles (user_id, dormitory_id, role_id) VALUES ($1, $2, $3)
			 ON CONFLICT (user_id, dormitory_id) DO UPDATE SET role_id = EXCLUDED.role_id, created_at = NOW()`,
			userID, item.DormitoryID, item.RoleID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO user_dormitories (user_id, dormitory_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			userID, item.DormitoryID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
