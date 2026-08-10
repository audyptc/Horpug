package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	coredomain "apigofiberhorpug/internal/domain"
	"apigofiberhorpug/internal/feature/expense/domain"

	"github.com/jackc/pgx/v5"
)

type ExpenseRepo struct {
	db *database.DB
}

func NewExpenseRepo(db *database.DB) *ExpenseRepo {
	return &ExpenseRepo{db: db}
}

func (r *ExpenseRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Expense, error) {
	e := &domain.Expense{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, dormitory_id, expense_date, category, description, amount, note, created_at, updated_at
		FROM expenses WHERE id = $1 AND dormitory_id = $2 AND deleted_at IS NULL`, id, dormitoryID).
		Scan(&e.ID, &e.DormitoryID, &e.ExpenseDate, &e.Category, &e.Description, &e.Amount, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("expense not found: %w", coredomain.ErrNotFound)
	}
	return e, err
}

func (r *ExpenseRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Expense, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, dormitory_id, expense_date, category, description, amount, note, created_at, updated_at
		FROM expenses
		WHERE dormitory_id = $1 AND deleted_at IS NULL
		ORDER BY expense_date DESC, created_at DESC
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Expense
	for rows.Next() {
		e := &domain.Expense{}
		if err := rows.Scan(&e.ID, &e.DormitoryID, &e.ExpenseDate, &e.Category, &e.Description, &e.Amount, &e.Note, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.Expense{}
	}
	return list, rows.Err()
}

func (r *ExpenseRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM expenses WHERE dormitory_id = $1 AND deleted_at IS NULL`, dormitoryID).Scan(&total)
	return total, err
}

func (r *ExpenseRepo) Create(ctx context.Context, e *domain.Expense) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO expenses (id, dormitory_id, expense_date, category, description, amount, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.DormitoryID, e.ExpenseDate, e.Category, e.Description, e.Amount, e.Note)
	return err
}

func (r *ExpenseRepo) Update(ctx context.Context, e *domain.Expense) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE expenses
		SET expense_date=$3, category=$4, description=$5, amount=$6, note=$7, updated_at=NOW()
		WHERE id=$1 AND dormitory_id=$2`,
		e.ID, e.DormitoryID, e.ExpenseDate, e.Category, e.Description, e.Amount, e.Note)
	return err
}

func (r *ExpenseRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE expenses SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND dormitory_id = $2 AND deleted_at IS NULL`,
		id, dormitoryID)
	return err
}
