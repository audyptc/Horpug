package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	roledomain "apihorpug/internal/features/role/domain"
	userdomain "apihorpug/internal/features/user/domain"
	userusecase "apihorpug/internal/features/user/usecase"

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

// buildScope builds the WHERE conditions shared by Count and List. Both must
// apply exactly the same conditions, otherwise the reported total disagrees
// with the rows returned and the client paginates over a page count that
// doesn't exist.
func (r *Repository) buildScope(filters userusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf(`u.is_active = $%d`, *argIdx))
		*args = append(*args, *filters.IsActive)
		*argIdx++
	}
	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			`(u.username ILIKE $%d OR u.email ILIKE $%d OR r.name ILIKE $%d)`,
			*argIdx, *argIdx, *argIdx,
		))
		*args = append(*args, "%"+filters.Search+"%")
		*argIdx++
	}

	// Sorted so the generated SQL is stable for a given set of filters rather
	// than varying with Go's randomised map iteration order.
	keys := make([]string, 0, len(filters.Columns))
	for key := range filters.Columns {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		column, ok := userusecase.FilterColumns[key]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", column, *argIdx))
		*args = append(*args, "%"+filters.Columns[key]+"%")
		*argIdx++
	}

	return conditions
}

// listOrderBy resolves the sort key through the usecase whitelist; anything
// unrecognised falls back to the default rather than reaching SQL. u.id
// breaks ties so paging over equal values can't repeat or skip a row.
func listOrderBy(filters userusecase.ListFilters) string {
	column, ok := userusecase.SortColumns[filters.SortKey]
	if !ok {
		column = userusecase.SortColumns[userusecase.DefaultSortKey]
	}

	direction := "ASC"
	if filters.SortDesc {
		direction = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s, u.id ASC", column, direction)
}

func (r *Repository) Count(ctx context.Context, filters userusecase.ListFilters) (int64, error) {
	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(filters, &argIdx, &args)

	query := `SELECT COUNT(*) FROM users u JOIN roles r ON r.id = u.role_id`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, filters userusecase.ListFilters, limit, offset int) ([]userdomain.User, error) {
	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(filters, &argIdx, &args)

	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
			u.is_protected,
			u.created_by,
			u.updated_by,
			u.created_at,
			u.updated_at,
			r.id,
			r.name,
			r.description,
			r.is_active,
			r.created_at,
			r.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
	`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += listOrderBy(filters) + fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]userdomain.User, 0)
	for rows.Next() {
		var user userdomain.User
		var role roledomain.Role
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.RoleID,
			&user.IsActive,
			&user.IsProtected,
			&user.CreatedBy,
			&user.UpdatedBy,
			&user.CreatedAt,
			&user.UpdatedAt,
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.Role = &role
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) ListActive(ctx context.Context, search string, limit int) ([]userdomain.User, error) {
	query := `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
			u.is_protected,
			u.created_by,
			u.updated_by,
			u.created_at,
			u.updated_at,
			r.id,
			r.name,
			r.description,
			r.is_active,
			r.created_at,
			r.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.is_active = true
	`
	args := make([]any, 0)
	argIdx := 1
	if search != "" {
		query += fmt.Sprintf(` AND (u.username ILIKE $%d OR u.email ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	query += fmt.Sprintf(` ORDER BY u.username ASC LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]userdomain.User, 0)
	for rows.Next() {
		var user userdomain.User
		var role roledomain.Role
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password,
			&user.RoleID,
			&user.IsActive,
			&user.IsProtected,
			&user.CreatedBy,
			&user.UpdatedBy,
			&user.CreatedAt,
			&user.UpdatedAt,
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.Role = &role
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *Repository) FindByLogin(ctx context.Context, login string) (userdomain.User, error) {
	var user userdomain.User
	var role roledomain.Role

	err := r.db.QueryRow(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
			u.is_protected,
			u.created_by,
			u.updated_by,
			u.created_at,
			u.updated_at,
			r.id,
			r.name,
			r.description,
			r.is_active,
			r.created_at,
			r.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.username = $1 OR u.email = $1
	`, login).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.IsActive,
		&user.IsProtected,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrUserNotFound
		}
		return userdomain.User{}, err
	}

	user.Role = &role
	return user, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (userdomain.User, error) {
	user, err := r.loadUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.User{}, userdomain.ErrUserNotFound
		}
		return userdomain.User{}, err
	}
	return user, nil
}

func (r *Repository) GetPermissions(ctx context.Context, id uuid.UUID) ([]userusecase.UserPermissionItem, error) {
	var roleID uuid.UUID
	if err := r.db.QueryRow(ctx, `SELECT role_id FROM users WHERE id = $1`, id).Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, userdomain.ErrUserNotFound
		}
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT rmp.menu_id, m.name AS menu_name, m.path AS menu_path, rmp.permission_id, p.name AS permission_name
		FROM role_menu_permissions rmp
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE rmp.role_id = $1
		ORDER BY m.path ASC, p.name ASC
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	permissions := make([]userusecase.UserPermissionItem, 0)
	for rows.Next() {
		var item userusecase.UserPermissionItem
		if err := rows.Scan(&item.MenuID, &item.MenuName, &item.MenuPath, &item.PermissionID, &item.PermissionName); err != nil {
			return nil, err
		}
		permissions = append(permissions, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}

// CountReferences counts rows across the tables that track who created or
// last updated them (created_by/updated_by) referencing this user, so a
// delete can be blocked in favor of deactivation when the user has audit
// history worth preserving.
func (r *Repository) CountReferences(ctx context.Context, id uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM room_types WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM rooms WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM tenants WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM contracts WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM electricity_meters WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM water_meters WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM invoices WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM payments WHERE created_by = $1) +
			(SELECT COUNT(*) FROM expenses WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM repair_requests WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM parking_registrations WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM parcels WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM announcements WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM documents WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM dormitories WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM roles WHERE created_by = $1 OR updated_by = $1) +
			(SELECT COUNT(*) FROM users WHERE created_by = $1 OR updated_by = $1)
	`, id).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) Create(ctx context.Context, input userusecase.CreateInput, hashedPassword string) (userdomain.User, error) {
	if err := r.ensureRoleExists(ctx, input.RoleID); err != nil {
		return userdomain.User{}, err
	}

	user := userdomain.User{
		ID:        uuid.New(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  hashedPassword,
		RoleID:    input.RoleID,
		IsActive:  input.IsActive,
		CreatedBy: input.CreatedBy,
		UpdatedBy: input.CreatedBy,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO users (id, username, email, password, role_id, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`, user.ID, user.Username, user.Email, user.Password, user.RoleID, user.IsActive, user.CreatedBy, user.UpdatedBy).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return userdomain.User{}, userdomain.ErrUserDuplicate
		}
		return userdomain.User{}, err
	}

	return r.loadUserByID(ctx, user.ID)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input userusecase.UpdateInput, hashedPassword *string) (userdomain.User, error) {
	if err := r.ensureUserExists(ctx, id); err != nil {
		return userdomain.User{}, err
	}

	if input.RoleID != nil {
		if err := r.ensureRoleExists(ctx, *input.RoleID); err != nil {
			return userdomain.User{}, err
		}
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, *input.Username)
		argIdx++
	}
	if input.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *input.Email)
		argIdx++
	}
	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *input.IsActive)
		argIdx++
	}
	if input.RoleID != nil {
		setClauses = append(setClauses, fmt.Sprintf("role_id = $%d", argIdx))
		args = append(args, *input.RoleID)
		argIdx++
	}
	if hashedPassword != nil {
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, *hashedPassword)
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
		query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return userdomain.User{}, userdomain.ErrUserDuplicate
			}
			return userdomain.User{}, err
		}
	}

	return r.loadUserByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return userdomain.ErrUserNotFound
	}

	return nil
}

func (r *Repository) ensureRoleExists(ctx context.Context, roleID uuid.UUID) error {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM roles WHERE id = $1`, roleID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.ErrRoleNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) ensureUserExists(ctx context.Context, userID uuid.UUID) error {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM users WHERE id = $1`, userID).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return userdomain.ErrUserNotFound
		}
		return err
	}
	return nil
}

func (r *Repository) loadUserByID(ctx context.Context, userID uuid.UUID) (userdomain.User, error) {
	var user userdomain.User
	var role roledomain.Role

	err := r.db.QueryRow(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
			u.is_protected,
			u.created_by,
			u.updated_by,
			u.created_at,
			u.updated_at,
			r.id,
			r.name,
			r.description,
			r.is_active,
			r.created_at,
			r.updated_at
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.RoleID,
		&user.IsActive,
		&user.IsProtected,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
		&role.ID,
		&role.Name,
		&role.Description,
		&role.IsActive,
		&role.CreatedAt,
		&role.UpdatedAt,
	)
	if err != nil {
		return userdomain.User{}, err
	}

	user.Role = &role
	return user, nil
}
