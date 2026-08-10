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

func (r *ExpenseRepo) FindByID(ctx context.Context, id string) (*domain.Expense, error) {
	e := &domain.Expense{}
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, expense_date, category, description, amount, note, created_at, updated_at
		FROM expenses WHERE id = $1 AND deleted_at IS NULL`, id).
		Scan(&e.ID, &e.ExpenseDate, &e.Category, &e.Description, &e.Amount, &e.Note, &e.CreatedAt, &e.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("expense not found: %w", coredomain.ErrNotFound)
	}
	return e, err
}

func (r *ExpenseRepo) List(ctx context.Context, limit, offset int) ([]*domain.Expense, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, expense_date, category, description, amount, note, created_at, updated_at
		FROM expenses
		WHERE deleted_at IS NULL
		ORDER BY expense_date DESC, created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Expense
	for rows.Next() {
		e := &domain.Expense{}
		if err := rows.Scan(&e.ID, &e.ExpenseDate, &e.Category, &e.Description, &e.Amount, &e.Note, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	if list == nil {
		list = []*domain.Expense{}
	}
	return list, rows.Err()
}

func (r *ExpenseRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM expenses WHERE deleted_at IS NULL`).Scan(&total)
	return total, err
}

func (r *ExpenseRepo) Create(ctx context.Context, e *domain.Expense) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO expenses (id, expense_date, category, description, amount, note)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, e.ExpenseDate, e.Category, e.Description, e.Amount, e.Note)
	return err
}

func (r *ExpenseRepo) Update(ctx context.Context, e *domain.Expense) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE expenses
		SET expense_date=$2, category=$3, description=$4, amount=$5, note=$6, updated_at=NOW()
		WHERE id=$1`,
		e.ID, e.ExpenseDate, e.Category, e.Description, e.Amount, e.Note)
	return err
}

func (r *ExpenseRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx,
		`UPDATE expenses SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
