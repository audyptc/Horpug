package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	expensedomain "apihorpug/internal/features/expense/domain"
	expenseusecase "apihorpug/internal/features/expense/usecase"

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

const selectExpenseColumns = `
	e.id, e.dormitory_id, d.name, e.category, e.expense_date, e.amount, e.description,
	e.created_by, e.updated_by, e.created_at, e.updated_at
`

const expenseFromJoins = `
	FROM expenses e
	JOIN dormitories d ON d.id = e.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters expenseusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`e.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`e.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf(`e.category = $%d`, *argIdx))
		*args = append(*args, *filters.Category)
		*argIdx++
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf(`e.expense_date >= $%d`, *argIdx))
		*args = append(*args, *filters.DateFrom)
		*argIdx++
	}
	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf(`e.expense_date <= $%d`, *argIdx))
		*args = append(*args, *filters.DateTo)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters expenseusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + expenseFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters expenseusecase.ListFilters, limit, offset int) ([]expensedomain.Expense, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectExpenseColumns + expenseFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY e.expense_date DESC, e.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanExpenses(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (expensedomain.Expense, error) {
	if err := r.ensureExpenseAccess(ctx, id, requesterID); err != nil {
		return expensedomain.Expense{}, err
	}

	expense, err := r.loadExpenseByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return expensedomain.Expense{}, expensedomain.ErrExpenseNotFound
		}
		return expensedomain.Expense{}, err
	}

	return expense, nil
}

func (r *Repository) Create(ctx context.Context, input expenseusecase.CreateInput) (expensedomain.Expense, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return expensedomain.Expense{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO expenses (id, dormitory_id, category, expense_date, amount, description, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, input.DormitoryID, input.Category, input.ExpenseDate, input.Amount, input.Description, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return expensedomain.Expense{}, expensedomain.ErrDormitoryNotFound
		}
		return expensedomain.Expense{}, err
	}

	return r.loadExpenseByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input expenseusecase.UpdateInput) (expensedomain.Expense, error) {
	if err := r.ensureExpenseAccess(ctx, id, requesterID); err != nil {
		return expensedomain.Expense{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *input.Category)
		argIdx++
	}
	if input.ExpenseDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("expense_date = $%d", argIdx))
		args = append(args, *input.ExpenseDate)
		argIdx++
	}
	if input.Amount != nil {
		setClauses = append(setClauses, fmt.Sprintf("amount = $%d", argIdx))
		args = append(args, *input.Amount)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
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
		query := fmt.Sprintf("UPDATE expenses SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return expensedomain.Expense{}, err
		}
	}

	return r.loadExpenseByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureExpenseAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return expensedomain.ErrExpenseNotFound
	}

	return nil
}

func (r *Repository) loadExpenseByID(ctx context.Context, id uuid.UUID) (expensedomain.Expense, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectExpenseColumns+expenseFromJoins+` WHERE e.id = $1`, id)
	return scanExpense(row)
}

func scanExpense(row pgx.Row) (expensedomain.Expense, error) {
	var expense expensedomain.Expense
	if err := row.Scan(
		&expense.ID,
		&expense.DormitoryID,
		&expense.DormitoryName,
		&expense.Category,
		&expense.ExpenseDate,
		&expense.Amount,
		&expense.Description,
		&expense.CreatedBy,
		&expense.UpdatedBy,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	); err != nil {
		return expensedomain.Expense{}, err
	}
	return expense, nil
}

func scanExpenses(rows pgx.Rows) ([]expensedomain.Expense, error) {
	expenses := make([]expensedomain.Expense, 0)
	for rows.Next() {
		expense, err := scanExpense(rows)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, expense)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return expenses, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages expenses in every dormitory), along with their
// role ID so callers can also check role-level dormitory grants.
func (r *Repository) dormitoryScope(ctx context.Context, userID uuid.UUID) (full bool, roleID uuid.UUID, err error) {
	err = r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&full, &roleID)
	if err != nil {
		return false, uuid.Nil, err
	}
	return full, roleID, nil
}

// ensureDormitoryAccess confirms the dormitory exists and the requester may
// record expenses under it (unrestricted, individually assigned via
// user_dormitories, or granted through their role via role_dormitories).
func (r *Repository) ensureDormitoryAccess(ctx context.Context, dormitoryID, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM dormitories d
		WHERE d.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = d.id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = d.id AND rd.role_id = $4
		))
	`, dormitoryID, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return expensedomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureExpenseAccess confirms the expense exists and the requester may act
// on it, based on access to its parent dormitory. Both a missing expense and
// a missing grant surface as ErrExpenseNotFound so scoped-out callers can't
// distinguish the two.
func (r *Repository) ensureExpenseAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM expenses e
		WHERE e.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = e.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = e.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return expensedomain.ErrExpenseNotFound
		}
		return err
	}
	return nil
}
