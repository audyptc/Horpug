package http

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	roledomain "apihorpug/internal/features/role/domain"
	userdomain "apihorpug/internal/features/user/domain"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db *pgxpool.Pool
}

type createUserRequest struct {
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
	RoleID   uuid.UUID `json:"role_id"`
	IsActive *bool     `json:"is_active"`
}

type updateUserRequest struct {
	Username *string    `json:"username"`
	Email    *string    `json:"email"`
	Password *string    `json:"password"`
	RoleID   *uuid.UUID `json:"role_id"`
	IsActive *bool      `json:"is_active"`
}

type userPermissionItem struct {
	MenuID         uuid.UUID `json:"menu_id"`
	MenuName       string    `json:"menu_name"`
	MenuPath       string    `json:"menu_path"`
	PermissionID   uuid.UUID `json:"permission_id"`
	PermissionName string    `json:"permission_name"`
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db}
}

func (h *Handler) List(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	rows, err := h.db.Query(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
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
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
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
			&user.CreatedAt,
			&user.UpdatedAt,
			&role.ID,
			&role.Name,
			&role.Description,
			&role.IsActive,
			&role.CreatedAt,
			&role.UpdatedAt,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
		}
		user.Role = &role
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list users"})
	}

	return c.JSON(users)
}

func (h *Handler) Get(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	user, err := loadUserByID(ctx, h.db, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get user"})
	}

	return c.JSON(user)
}

func (h *Handler) GetPermissions(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	var roleID uuid.UUID
	if err := h.db.QueryRow(ctx, `SELECT role_id FROM users WHERE id = $1`, id).Scan(&roleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get user"})
	}

	rows, err := h.db.Query(ctx, `
		SELECT rmp.menu_id, m.name AS menu_name, m.path AS menu_path, rmp.permission_id, p.name AS permission_name
		FROM role_menu_permissions rmp
		JOIN menus m ON m.id = rmp.menu_id
		JOIN permissions p ON p.id = rmp.permission_id
		WHERE rmp.role_id = $1
		ORDER BY m.path ASC, p.name ASC
	`, roleID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user permissions"})
	}
	defer rows.Close()

	permissions := make([]userPermissionItem, 0)
	for rows.Next() {
		var item userPermissionItem
		if err := rows.Scan(&item.MenuID, &item.MenuName, &item.MenuPath, &item.PermissionID, &item.PermissionName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user permissions"})
		}
		permissions = append(permissions, item)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user permissions"})
	}

	return c.JSON(permissions)
}

func (h *Handler) Create(c fiber.Ctx) error {
	var req createUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username, email and password are required"})
	}
	if req.RoleID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role_id is required"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := ensureRoleExists(ctx, h.db, req.RoleID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to validate role"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
	}

	user := userdomain.User{
		ID:       uuid.New(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		RoleID:   req.RoleID,
		IsActive: true,
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	err = h.db.QueryRow(ctx, `
		INSERT INTO users (id, username, email, password, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, user.ID, user.Username, user.Email, user.Password, user.RoleID, user.IsActive).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username or email already exists"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create user"})
	}

	createdUser, err := loadUserByID(ctx, h.db, user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user"})
	}

	return c.Status(fiber.StatusCreated).JSON(createdUser)
}

func (h *Handler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	var req updateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	if err := ensureUserExists(ctx, h.db, id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get user"})
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		if username == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "username cannot be empty"})
		}
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, username)
		argIdx++
	}
	if req.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*req.Email))
		if email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "email cannot be empty"})
		}
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, email)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.RoleID != nil {
		if err := ensureRoleExists(ctx, h.db, *req.RoleID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to validate role"})
		}
		setClauses = append(setClauses, fmt.Sprintf("role_id = $%d", argIdx))
		args = append(args, *req.RoleID)
		argIdx++
	}
	if req.Password != nil {
		if strings.TrimSpace(*req.Password) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "password cannot be empty"})
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to hash password"})
		}
		setClauses = append(setClauses, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, string(hashedPassword))
		argIdx++
	}

	if len(setClauses) > 0 {
		setClauses = append(setClauses, "updated_at = NOW()")
		args = append(args, id)
		query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := h.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "username or email already exists"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to update user"})
		}
	}

	updatedUser, err := loadUserByID(ctx, h.db, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to load user"})
	}

	return c.JSON(updatedUser)
}

func (h *Handler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	result, err := h.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete user"})
	}
	if result.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{"message": "user deleted"})
}

func ensureRoleExists(ctx context.Context, db *pgxpool.Pool, roleID uuid.UUID) error {
	var exists int
	return db.QueryRow(ctx, `SELECT 1 FROM roles WHERE id = $1`, roleID).Scan(&exists)
}

func ensureUserExists(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) error {
	var exists int
	return db.QueryRow(ctx, `SELECT 1 FROM users WHERE id = $1`, userID).Scan(&exists)
}

func loadUserByID(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (userdomain.User, error) {
	var user userdomain.User
	var role roledomain.Role

	err := db.QueryRow(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.password,
			u.role_id,
			u.is_active,
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
