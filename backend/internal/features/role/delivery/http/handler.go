package http

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	roledomain "apihorpug/internal/features/role/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errReferenceNotFound = errors.New("one or more menus or permissions not found")

type Handler struct {
	db *pgxpool.Pool
}

type menuPermissionInput struct {
	MenuID        uuid.UUID   `json:"menu_id"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type createRoleRequest struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	IsActive        *bool                 `json:"is_active"`
	MenuPermissions []menuPermissionInput `json:"menu_permissions"`
}

type updateRoleRequest struct {
	Name            *string                `json:"name"`
	Description     *string                `json:"description"`
	IsActive        *bool                  `json:"is_active"`
	MenuPermissions *[]menuPermissionInput `json:"menu_permissions"`
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM roles
		ORDER BY name ASC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
	}
	defer rows.Close()

	roles := make([]roledomain.Role, 0)
	for rows.Next() {
		var role roledomain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
		}

		menuPermissions, err := fetchRoleMenuPermissions(ctx, h.db, role.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
		}
		role.MenuPermissions = menuPermissions
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list roles"})
	}

	return c.JSON(roles)
}

func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	role, err := loadRoleByID(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get role"})
	}

	return c.JSON(role)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req createRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name is required"})
	}

	role := roledomain.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		IsActive:    true,
	}
	if req.IsActive != nil {
		role.IsActive = *req.IsActive
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create role"})
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx, `
		INSERT INTO roles (id, name, description, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at
	`, role.ID, role.Name, role.Description, role.IsActive).Scan(&role.CreatedAt, &role.UpdatedAt)
	if err == nil {
		err = replaceRoleMenuPermissions(ctx, tx, role.ID, req.MenuPermissions)
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
		}
		if errors.Is(err, errReferenceNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create role"})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create role"})
	}

	createdRole, err := loadRoleByID(ctx, h.db, role.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load role"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdRole)
}

func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	var req updateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var exists int
	if err := h.db.QueryRow(ctx, `SELECT 1 FROM roles WHERE id = $1`, id).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get role"})
	}

	tx, err := h.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
	}
	defer tx.Rollback(ctx)

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name cannot be empty"})
		}
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, strings.TrimSpace(*req.Description))
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = NOW()")
		args = append(args, id)
		query := fmt.Sprintf("UPDATE roles SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := tx.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role name already exists"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
		}
	}

	if req.MenuPermissions != nil {
		if err := replaceRoleMenuPermissions(ctx, tx, id, *req.MenuPermissions); err != nil {
			if errors.Is(err, errReferenceNotFound) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "one or more menus or permissions not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update role"})
	}

	updatedRole, err := loadRoleByID(ctx, h.db, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load role"})
	}

	return c.JSON(updatedRole)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid role id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	result, err := h.db.Exec(ctx, `DELETE FROM roles WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "role is being used by users"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete role"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "role not found"})
	}

	return c.JSON(fiber.Map{"message": "role deleted"})
}

func loadRoleByID(ctx context.Context, db *pgxpool.Pool, roleID uuid.UUID) (roledomain.Role, error) {
	var role roledomain.Role
	err := db.QueryRow(ctx, `
		SELECT id, name, description, is_active, created_at, updated_at
		FROM roles
		WHERE id = $1
	`, roleID).Scan(&role.ID, &role.Name, &role.Description, &role.IsActive, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return roledomain.Role{}, err
	}

	menuPermissions, err := fetchRoleMenuPermissions(ctx, db, role.ID)
	if err != nil {
		return roledomain.Role{}, err
	}
	role.MenuPermissions = menuPermissions
	return role, nil
}

func fetchRoleMenuPermissions(ctx context.Context, db *pgxpool.Pool, roleID uuid.UUID) ([]roledomain.RoleMenuPermission, error) {
	rows, err := db.Query(ctx, `
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

func replaceRoleMenuPermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID, menuPermissions []menuPermissionInput) error {
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
			return errReferenceNotFound
		}
		menuSet[mp.MenuID] = struct{}{}
		for _, permissionID := range mp.PermissionIDs {
			if permissionID == uuid.Nil {
				return errReferenceNotFound
			}
			permSet[permissionID] = struct{}{}
			rows = append(rows, roledomain.RoleMenuPermission{RoleID: roleID, MenuID: mp.MenuID, PermissionID: permissionID})
		}
	}

	for menuID := range menuSet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM menus WHERE id = $1`, menuID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errReferenceNotFound
			}
			return err
		}
	}

	for permissionID := range permSet {
		if err := tx.QueryRow(ctx, `SELECT 1 FROM permissions WHERE id = $1`, permissionID).Scan(new(int)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errReferenceNotFound
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
