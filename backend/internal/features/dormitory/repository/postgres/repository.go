package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	dormdomain "apihorpug/internal/features/dormitory/domain"
	dormusecase "apihorpug/internal/features/dormitory/usecase"

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

func (r *Repository) List(ctx context.Context, filter dormusecase.ListFilter) ([]dormdomain.Dormitory, error) {
	query := `
		SELECT DISTINCT d.id, d.name, d.address, d.phone, d.description, d.is_active, d.created_by, d.updated_by, d.created_at, d.updated_at
		FROM dormitories d
	`
	args := make([]any, 0)
	if filter.UserID != nil {
		query += ` JOIN user_dormitories ud ON ud.dormitory_id = d.id WHERE ud.user_id = $1`
		args = append(args, *filter.UserID)
	}
	query += ` ORDER BY d.name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dormitories := make([]dormdomain.Dormitory, 0)
	for rows.Next() {
		var dormitory dormdomain.Dormitory
		if err := rows.Scan(
			&dormitory.ID,
			&dormitory.Name,
			&dormitory.Address,
			&dormitory.Phone,
			&dormitory.Description,
			&dormitory.IsActive,
			&dormitory.CreatedBy,
			&dormitory.UpdatedBy,
			&dormitory.CreatedAt,
			&dormitory.UpdatedAt,
		); err != nil {
			return nil, err
		}

		managers, err := r.fetchManagers(ctx, dormitory.ID)
		if err != nil {
			return nil, err
		}
		dormitory.Managers = managers
		dormitories = append(dormitories, dormitory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dormitories, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (dormdomain.Dormitory, error) {
	dormitory, err := r.loadDormitoryByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dormdomain.Dormitory{}, dormdomain.ErrDormitoryNotFound
		}
		return dormdomain.Dormitory{}, err
	}
	return dormitory, nil
}

func (r *Repository) Create(ctx context.Context, input dormusecase.CreateInput) (dormdomain.Dormitory, error) {
	dormitory := dormdomain.Dormitory{
		ID:          uuid.New(),
		Name:        input.Name,
		Address:     input.Address,
		Phone:       input.Phone,
		Description: input.Description,
		IsActive:    input.IsActive,
		CreatedBy:   input.CreatedBy,
		UpdatedBy:   input.CreatedBy,
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return dormdomain.Dormitory{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO dormitories (id, name, address, phone, description, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`, dormitory.ID, dormitory.Name, dormitory.Address, dormitory.Phone, dormitory.Description, dormitory.IsActive, dormitory.CreatedBy, dormitory.UpdatedBy).
		Scan(&dormitory.CreatedAt, &dormitory.UpdatedAt)
	if err == nil {
		err = r.replaceManagers(ctx, tx, dormitory.ID, input.ManagerIDs)
	}
	if err != nil {
		return dormdomain.Dormitory{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dormdomain.Dormitory{}, err
	}

	return r.loadDormitoryByID(ctx, dormitory.ID)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input dormusecase.UpdateInput) (dormdomain.Dormitory, error) {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM dormitories WHERE id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dormdomain.Dormitory{}, dormdomain.ErrDormitoryNotFound
		}
		return dormdomain.Dormitory{}, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return dormdomain.Dormitory{}, err
	}
	defer tx.Rollback(ctx)

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Address != nil {
		setClauses = append(setClauses, fmt.Sprintf("address = $%d", argIdx))
		args = append(args, *input.Address)
		argIdx++
	}
	if input.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *input.Phone)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
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
		query := fmt.Sprintf("UPDATE dormitories SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			return dormdomain.Dormitory{}, err
		}
	}

	if input.ManagerIDs != nil {
		if err := r.replaceManagers(ctx, tx, id, *input.ManagerIDs); err != nil {
			return dormdomain.Dormitory{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return dormdomain.Dormitory{}, err
	}

	return r.loadDormitoryByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM dormitories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return dormdomain.ErrDormitoryNotFound
	}

	return nil
}

func (r *Repository) loadDormitoryByID(ctx context.Context, id uuid.UUID) (dormdomain.Dormitory, error) {
	var dormitory dormdomain.Dormitory
	err := r.db.QueryRow(ctx, `
		SELECT id, name, address, phone, description, is_active, created_by, updated_by, created_at, updated_at
		FROM dormitories
		WHERE id = $1
	`, id).Scan(
		&dormitory.ID,
		&dormitory.Name,
		&dormitory.Address,
		&dormitory.Phone,
		&dormitory.Description,
		&dormitory.IsActive,
		&dormitory.CreatedBy,
		&dormitory.UpdatedBy,
		&dormitory.CreatedAt,
		&dormitory.UpdatedAt,
	)
	if err != nil {
		return dormdomain.Dormitory{}, err
	}

	managers, err := r.fetchManagers(ctx, dormitory.ID)
	if err != nil {
		return dormdomain.Dormitory{}, err
	}
	dormitory.Managers = managers

	return dormitory, nil
}

func (r *Repository) fetchManagers(ctx context.Context, dormitoryID uuid.UUID) ([]dormdomain.DormitoryManager, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.id, u.username, u.email, ud.created_at
		FROM user_dormitories ud
		JOIN users u ON u.id = ud.user_id
		WHERE ud.dormitory_id = $1
		ORDER BY u.username ASC
	`, dormitoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	managers := make([]dormdomain.DormitoryManager, 0)
	for rows.Next() {
		var manager dormdomain.DormitoryManager
		if err := rows.Scan(&manager.UserID, &manager.Username, &manager.Email, &manager.CreatedAt); err != nil {
			return nil, err
		}
		managers = append(managers, manager)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return managers, nil
}

func (r *Repository) replaceManagers(ctx context.Context, tx pgx.Tx, dormitoryID uuid.UUID, userIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM user_dormitories WHERE dormitory_id = $1`, dormitoryID); err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}

	userSet := make(map[uuid.UUID]struct{}, len(userIDs))
	for _, userID := range userIDs {
		if userID == uuid.Nil {
			return dormdomain.ErrManagerNotFound
		}
		userSet[userID] = struct{}{}
	}

	for userID := range userSet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM users WHERE id = $1`, userID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return dormdomain.ErrManagerNotFound
			}
			return err
		}
	}

	for userID := range userSet {
		if _, err := tx.Exec(ctx, `
			INSERT INTO user_dormitories (user_id, dormitory_id)
			VALUES ($1, $2)
		`, userID, dormitoryID); err != nil {
			return err
		}
	}

	return nil
}
