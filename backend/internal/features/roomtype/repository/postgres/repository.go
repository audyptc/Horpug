package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	roomtypedomain "apihorpug/internal/features/roomtype/domain"
	roomtypeusecase "apihorpug/internal/features/roomtype/usecase"

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
func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters roomtypeusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rt.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rt.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf(`rt.is_active = $%d`, *argIdx))
		*args = append(*args, *filters.IsActive)
		*argIdx++
	}
	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			`(rt.name ILIKE $%d OR d.name ILIKE $%d)`,
			*argIdx, *argIdx,
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
		column, ok := roomtypeusecase.FilterColumns[key]
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
// unrecognised falls back to the default rather than reaching SQL. rt.id
// breaks ties so paging over equal values can't repeat or skip a row.
func listOrderBy(filters roomtypeusecase.ListFilters) string {
	column, ok := roomtypeusecase.SortColumns[filters.SortKey]
	if !ok {
		column = roomtypeusecase.SortColumns[roomtypeusecase.DefaultSortKey]
	}

	direction := "ASC"
	if filters.SortDesc {
		direction = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s, rt.id ASC", column, direction)
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters roomtypeusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) FROM room_types rt JOIN dormitories d ON d.id = rt.dormitory_id`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters roomtypeusecase.ListFilters, limit, offset int) ([]roomtypedomain.RoomType, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `
		SELECT rt.id, rt.dormitory_id, d.name, rt.name, rt.description, rt.price, rt.is_active, rt.created_by, rt.updated_by, rt.created_at, rt.updated_at
		FROM room_types rt
		JOIN dormitories d ON d.id = rt.dormitory_id
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

	roomTypes := make([]roomtypedomain.RoomType, 0)
	for rows.Next() {
		var roomType roomtypedomain.RoomType
		if err := rows.Scan(
			&roomType.ID,
			&roomType.DormitoryID,
			&roomType.DormitoryName,
			&roomType.Name,
			&roomType.Description,
			&roomType.Price,
			&roomType.IsActive,
			&roomType.CreatedBy,
			&roomType.UpdatedBy,
			&roomType.CreatedAt,
			&roomType.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roomTypes = append(roomTypes, roomType)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roomTypes, nil
}

func (r *Repository) ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, search string, limit int) ([]roomtypedomain.RoomType, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT rt.id, rt.dormitory_id, d.name, rt.name, rt.description, rt.price, rt.is_active, rt.created_by, rt.updated_by, rt.created_at, rt.updated_at
		FROM room_types rt
		JOIN dormitories d ON d.id = rt.dormitory_id
		WHERE rt.is_active = true
	`
	args := make([]any, 0)
	argIdx := 1

	if !full {
		query += fmt.Sprintf(` AND rt.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, argIdx, argIdx+1)
		args = append(args, requesterID, roleID)
		argIdx += 2
	}
	if dormitoryID != nil {
		query += fmt.Sprintf(` AND rt.dormitory_id = $%d`, argIdx)
		args = append(args, *dormitoryID)
		argIdx++
	}
	if search != "" {
		query += fmt.Sprintf(` AND rt.name ILIKE $%d`, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	query += fmt.Sprintf(` ORDER BY d.name ASC, rt.name ASC LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roomTypes := make([]roomtypedomain.RoomType, 0)
	for rows.Next() {
		var roomType roomtypedomain.RoomType
		if err := rows.Scan(
			&roomType.ID,
			&roomType.DormitoryID,
			&roomType.DormitoryName,
			&roomType.Name,
			&roomType.Description,
			&roomType.Price,
			&roomType.IsActive,
			&roomType.CreatedBy,
			&roomType.UpdatedBy,
			&roomType.CreatedAt,
			&roomType.UpdatedAt,
		); err != nil {
			return nil, err
		}
		roomTypes = append(roomTypes, roomType)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roomTypes, nil
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomtypedomain.RoomType, error) {
	if err := r.ensureRoomTypeAccess(ctx, id, requesterID); err != nil {
		return roomtypedomain.RoomType{}, err
	}

	roomType, err := r.loadRoomTypeByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roomtypedomain.RoomType{}, roomtypedomain.ErrRoomTypeNotFound
		}
		return roomtypedomain.RoomType{}, err
	}
	return roomType, nil
}

func (r *Repository) Create(ctx context.Context, input roomtypeusecase.CreateInput) (roomtypedomain.RoomType, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return roomtypedomain.RoomType{}, err
		}
	}

	roomType := roomtypedomain.RoomType{
		ID:          uuid.New(),
		DormitoryID: input.DormitoryID,
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		IsActive:    input.IsActive,
		CreatedBy:   input.CreatedBy,
		UpdatedBy:   input.CreatedBy,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO room_types (id, dormitory_id, name, description, price, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`, roomType.ID, roomType.DormitoryID, roomType.Name, roomType.Description, roomType.Price, roomType.IsActive, roomType.CreatedBy, roomType.UpdatedBy).
		Scan(&roomType.CreatedAt, &roomType.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return roomtypedomain.RoomType{}, roomtypedomain.ErrRoomTypeNameExists
			}
			if pgErr.Code == "23503" {
				return roomtypedomain.RoomType{}, roomtypedomain.ErrDormitoryNotFound
			}
		}
		return roomtypedomain.RoomType{}, err
	}

	return r.loadRoomTypeByID(ctx, roomType.ID)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input roomtypeusecase.UpdateInput) (roomtypedomain.RoomType, error) {
	if err := r.ensureRoomTypeAccess(ctx, id, requesterID); err != nil {
		return roomtypedomain.RoomType{}, err
	}

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
	if input.Price != nil {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIdx))
		args = append(args, *input.Price)
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
		query := fmt.Sprintf("UPDATE room_types SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return roomtypedomain.RoomType{}, roomtypedomain.ErrRoomTypeNameExists
			}
			return roomtypedomain.RoomType{}, err
		}
	}

	return r.loadRoomTypeByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureRoomTypeAccess(ctx, id, requesterID); err != nil {
		return err
	}

	roomCount, err := r.CountRooms(ctx, id)
	if err != nil {
		return err
	}
	if roomCount > 0 {
		return roomtypedomain.ErrRoomTypeHasRooms
	}

	result, err := r.db.Exec(ctx, `DELETE FROM room_types WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return roomtypedomain.ErrRoomTypeNotFound
	}

	return nil
}

func (r *Repository) CountRooms(ctx context.Context, id uuid.UUID) (int64, error) {
	var roomCount int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rooms WHERE room_type_id = $1`, id).Scan(&roomCount)
	return roomCount, err
}

func (r *Repository) loadRoomTypeByID(ctx context.Context, id uuid.UUID) (roomtypedomain.RoomType, error) {
	var roomType roomtypedomain.RoomType
	err := r.db.QueryRow(ctx, `
		SELECT rt.id, rt.dormitory_id, d.name, rt.name, rt.description, rt.price, rt.is_active, rt.created_by, rt.updated_by, rt.created_at, rt.updated_at
		FROM room_types rt
		JOIN dormitories d ON d.id = rt.dormitory_id
		WHERE rt.id = $1
	`, id).Scan(
		&roomType.ID,
		&roomType.DormitoryID,
		&roomType.DormitoryName,
		&roomType.Name,
		&roomType.Description,
		&roomType.Price,
		&roomType.IsActive,
		&roomType.CreatedBy,
		&roomType.UpdatedBy,
		&roomType.CreatedAt,
		&roomType.UpdatedAt,
	)
	if err != nil {
		return roomtypedomain.RoomType{}, err
	}
	return roomType, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages room types in every dormitory), along with their
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
// manage room types under it (unrestricted, individually assigned via
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
			return roomtypedomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureRoomTypeAccess confirms the room type exists and the requester may
// act on it, based on access to its parent dormitory. Both a missing room
// type and a missing grant surface as ErrRoomTypeNotFound so scoped-out
// callers can't distinguish the two.
func (r *Repository) ensureRoomTypeAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM room_types rt
		WHERE rt.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rt.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rt.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roomtypedomain.ErrRoomTypeNotFound
		}
		return err
	}
	return nil
}
