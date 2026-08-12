package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	roledomain "apihorpug/internal/features/role/domain"
	roleusecase "apihorpug/internal/features/role/usecase"

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
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM roles`).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]roledomain.Role, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, name, description, is_active, full_dormitory_access, created_by, updated_by, created_at, updated_at
		FROM roles
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]roledomain.Role, 0)
	for rows.Next() {
		var role roledomain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.FullDormitoryAccess, &role.CreatedBy, &role.UpdatedBy, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}

		menuPermissions, err := r.fetchRoleMenuPermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.MenuPermissions = menuPermissions

		dormitories, err := r.fetchRoleDormitories(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		role.Dormitories = dormitories

		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (roledomain.Role, error) {
	role, err := r.loadRoleByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roledomain.Role{}, roledomain.ErrRoleNotFound
		}
		return roledomain.Role{}, err
	}
	return role, nil
}

func (r *Repository) Create(ctx context.Context, input roleusecase.CreateInput) (roledomain.Role, error) {
	role := roledomain.Role{
		ID:                  uuid.New(),
		Name:                input.Name,
		Description:         input.Description,
		IsActive:            input.IsActive,
		FullDormitoryAccess: input.FullDormitoryAccess,
		CreatedBy:           input.CreatedBy,
		UpdatedBy:           input.CreatedBy,
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return roledomain.Role{}, err
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO roles (id, name, description, is_active, full_dormitory_access, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`, role.ID, role.Name, role.Description, role.IsActive, role.FullDormitoryAccess, role.CreatedBy, role.UpdatedBy).Scan(&role.CreatedAt, &role.UpdatedAt)
	if err == nil {
		err = r.replaceRoleMenuPermissions(ctx, tx, role.ID, input.MenuPermissions)
	}
	if err == nil {
		err = r.replaceRoleDormitories(ctx, tx, role.ID, input.DormitoryIDs)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return roledomain.Role{}, roledomain.ErrRoleNameExists
		}
		if errors.Is(err, roledomain.ErrReferenceNotFound) {
			return roledomain.Role{}, err
		}
		return roledomain.Role{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return roledomain.Role{}, err
	}

	return r.loadRoleByID(ctx, role.ID)
}

func (r *Repository) Update(ctx context.Context, id uuid.UUID, input roleusecase.UpdateInput) (roledomain.Role, error) {
	var exists int
	if err := r.db.QueryRow(ctx, `SELECT 1 FROM roles WHERE id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roledomain.Role{}, roledomain.ErrRoleNotFound
		}
		return roledomain.Role{}, err
	}

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return roledomain.Role{}, err
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
	if input.FullDormitoryAccess != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_dormitory_access = $%d", argIdx))
		args = append(args, *input.FullDormitoryAccess)
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
		query := fmt.Sprintf("UPDATE roles SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return roledomain.Role{}, roledomain.ErrRoleNameExists
			}
			return roledomain.Role{}, err
		}
	}

	if input.MenuPermissions != nil {
		if err := r.replaceRoleMenuPermissions(ctx, tx, id, *input.MenuPermissions); err != nil {
			if errors.Is(err, roledomain.ErrReferenceNotFound) {
				return roledomain.Role{}, err
			}
			return roledomain.Role{}, err
		}
	}

	if input.DormitoryIDs != nil {
		if err := r.replaceRoleDormitories(ctx, tx, id, *input.DormitoryIDs); err != nil {
			return roledomain.Role{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return roledomain.Role{}, err
	}

	return r.loadRoleByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return roledomain.ErrRoleInUse
		}
		return err
	}
	if result.RowsAffected() == 0 {
		return roledomain.ErrRoleNotFound
	}

	return nil
}

func (r *Repository) loadRoleByID(ctx context.Context, roleID uuid.UUID) (roledomain.Role, error) {
	var role roledomain.Role
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, is_active, full_dormitory_access, created_by, updated_by, created_at, updated_at
		FROM roles
		WHERE id = $1
	`, roleID).Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.FullDormitoryAccess, &role.CreatedBy, &role.UpdatedBy, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return roledomain.Role{}, err
	}

	menuPermissions, err := r.fetchRoleMenuPermissions(ctx, role.ID)
	if err != nil {
		return roledomain.Role{}, err
	}
	role.MenuPermissions = menuPermissions

	dormitories, err := r.fetchRoleDormitories(ctx, role.ID)
	if err != nil {
		return roledomain.Role{}, err
	}
	role.Dormitories = dormitories

	return role, nil
}

func (r *Repository) fetchRoleMenuPermissions(ctx context.Context, roleID uuid.UUID) ([]roledomain.RoleMenuPermission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			rmp.role_id,
			rmp.menu_id,
			rmp.permission_id,
			rmp.created_at,
			m.id,
			m.name,
			m.path,
			m.description,
			m.is_active,
			m.created_at,
			m.updated_at,
			p.id,
			p.name,
			p.description,
			p.created_at,
			p.updated_at
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

	items := make([]roledomain.RoleMenuPermission, 0)
	for rows.Next() {
		var item roledomain.RoleMenuPermission
		if err := rows.Scan(
			&item.RoleID,
			&item.MenuID,
			&item.PermissionID,
			&item.CreatedAt,
			&item.Menu.ID,
			&item.Menu.Name,
			&item.Menu.Path,
			&item.Menu.Description,
			&item.Menu.IsActive,
			&item.Menu.CreatedAt,
			&item.Menu.UpdatedAt,
			&item.Permission.ID,
			&item.Permission.Name,
			&item.Permission.Description,
			&item.Permission.CreatedAt,
			&item.Permission.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) replaceRoleMenuPermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID, menuPermissions []roleusecase.MenuPermissionInput) error {
	if _, err := tx.Exec(ctx, `DELETE FROM role_menu_permissions WHERE role_id = $1`, roleID); err != nil {
		return err
	}
	if len(menuPermissions) == 0 {
		return nil
	}

	menuSet := make(map[uuid.UUID]struct{})
	permSet := make(map[uuid.UUID]struct{})
	rows := make([]roledomain.RoleMenuPermission, 0)
	for _, mp := range menuPermissions {
		if mp.MenuID == uuid.Nil || len(mp.PermissionIDs) == 0 {
			return roledomain.ErrReferenceNotFound
		}
		menuSet[mp.MenuID] = struct{}{}
		for _, permissionID := range mp.PermissionIDs {
			if permissionID == uuid.Nil {
				return roledomain.ErrReferenceNotFound
			}
			permSet[permissionID] = struct{}{}
			rows = append(rows, roledomain.RoleMenuPermission{RoleID: roleID, MenuID: mp.MenuID, PermissionID: permissionID})
		}
	}

	for menuID := range menuSet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM menus WHERE id = $1`, menuID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return roledomain.ErrReferenceNotFound
			}
			return err
		}
	}

	for permissionID := range permSet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM permissions WHERE id = $1`, permissionID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return roledomain.ErrReferenceNotFound
			}
			return err
		}
	}

	for _, row := range rows {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_menu_permissions (role_id, menu_id, permission_id)
			VALUES ($1, $2, $3)
		`, row.RoleID, row.MenuID, row.PermissionID); err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) fetchRoleDormitories(ctx context.Context, roleID uuid.UUID) ([]roledomain.RoleDormitory, error) {
	rows, err := r.db.Query(ctx, `
		SELECT d.id, d.name
		FROM role_dormitories rd
		JOIN dormitories d ON d.id = rd.dormitory_id
		WHERE rd.role_id = $1
		ORDER BY d.name ASC
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dormitories := make([]roledomain.RoleDormitory, 0)
	for rows.Next() {
		var dormitory roledomain.RoleDormitory
		if err := rows.Scan(&dormitory.ID, &dormitory.Name); err != nil {
			return nil, err
		}
		dormitories = append(dormitories, dormitory)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dormitories, nil
}

func (r *Repository) replaceRoleDormitories(ctx context.Context, tx pgx.Tx, roleID uuid.UUID, dormitoryIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM role_dormitories WHERE role_id = $1`, roleID); err != nil {
		return err
	}
	if len(dormitoryIDs) == 0 {
		return nil
	}

	dormitorySet := make(map[uuid.UUID]struct{}, len(dormitoryIDs))
	for _, dormitoryID := range dormitoryIDs {
		if dormitoryID == uuid.Nil {
			return roledomain.ErrReferenceNotFound
		}
		dormitorySet[dormitoryID] = struct{}{}
	}

	for dormitoryID := range dormitorySet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM dormitories WHERE id = $1`, dormitoryID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return roledomain.ErrReferenceNotFound
			}
			return err
		}
	}

	for dormitoryID := range dormitorySet {
		if _, err := tx.Exec(ctx, `
			INSERT INTO role_dormitories (role_id, dormitory_id)
			VALUES ($1, $2)
		`, roleID, dormitoryID); err != nil {
			return err
		}
	}

	return nil
}
