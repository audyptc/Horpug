package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	roomdomain "apihorpug/internal/features/room/domain"
	roomusecase "apihorpug/internal/features/room/usecase"

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

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	query := `SELECT COUNT(*) FROM rooms rm`
	conditions := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, argIdx, argIdx+1))
		args = append(args, requesterID, roleID)
		argIdx += 2
	}
	if dormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, argIdx))
		args = append(args, *dormitoryID)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, limit, offset int) ([]roomdomain.Room, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT rm.id, rm.dormitory_id, d.name, rm.room_type_id, rt.name, rt.price, rm.room_number, rm.floor, rm.status, rm.is_active, rm.created_by, rm.updated_by, rm.created_at, rm.updated_at
		FROM rooms rm
		JOIN dormitories d ON d.id = rm.dormitory_id
		JOIN room_types rt ON rt.id = rm.room_type_id
	`
	conditions := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, argIdx, argIdx+1))
		args = append(args, requesterID, roleID)
		argIdx += 2
	}
	if dormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, argIdx))
		args = append(args, *dormitoryID)
		argIdx++
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY d.name ASC, rm.room_number ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]roomdomain.Room, 0)
	for rows.Next() {
		var room roomdomain.Room
		if err := rows.Scan(
			&room.ID,
			&room.DormitoryID,
			&room.DormitoryName,
			&room.RoomTypeID,
			&room.RoomTypeName,
			&room.RoomTypePrice,
			&room.RoomNumber,
			&room.Floor,
			&room.Status,
			&room.IsActive,
			&room.CreatedBy,
			&room.UpdatedBy,
			&room.CreatedAt,
			&room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) ListActive(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, status *roomdomain.RoomStatus, search string, limit int) ([]roomdomain.Room, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT rm.id, rm.dormitory_id, d.name, rm.room_type_id, rt.name, rt.price, rm.room_number, rm.floor, rm.status, rm.is_active, rm.created_by, rm.updated_by, rm.created_at, rm.updated_at
		FROM rooms rm
		JOIN dormitories d ON d.id = rm.dormitory_id
		JOIN room_types rt ON rt.id = rm.room_type_id
		WHERE rm.is_active = true
	`
	args := make([]any, 0)
	argIdx := 1

	if !full {
		query += fmt.Sprintf(` AND rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, argIdx, argIdx+1)
		args = append(args, requesterID, roleID)
		argIdx += 2
	}
	if dormitoryID != nil {
		query += fmt.Sprintf(` AND rm.dormitory_id = $%d`, argIdx)
		args = append(args, *dormitoryID)
		argIdx++
	}
	if status != nil {
		query += fmt.Sprintf(` AND rm.status = $%d`, argIdx)
		args = append(args, *status)
		argIdx++

		// rm.status is set manually and can drift out of sync with reality, so
		// when filtering for available rooms also exclude any room that
		// already has an active contract — otherwise a stale "available"
		// status would let it be picked for a new contract and fail on the
		// one-active-contract-per-room constraint.
		if *status == roomdomain.RoomStatusAvailable {
			query += ` AND NOT EXISTS (SELECT 1 FROM contracts ct WHERE ct.room_id = rm.id AND ct.status = 'active')`
		}
	}
	if search != "" {
		query += fmt.Sprintf(` AND (rm.room_number ILIKE $%d OR d.name ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}
	query += fmt.Sprintf(` ORDER BY d.name ASC, rm.room_number ASC LIMIT $%d`, argIdx)
	args = append(args, limit)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]roomdomain.Room, 0)
	for rows.Next() {
		var room roomdomain.Room
		if err := rows.Scan(
			&room.ID,
			&room.DormitoryID,
			&room.DormitoryName,
			&room.RoomTypeID,
			&room.RoomTypeName,
			&room.RoomTypePrice,
			&room.RoomNumber,
			&room.Floor,
			&room.Status,
			&room.IsActive,
			&room.CreatedBy,
			&room.UpdatedBy,
			&room.CreatedAt,
			&room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rooms, nil
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (roomdomain.Room, error) {
	if err := r.ensureRoomAccess(ctx, id, requesterID); err != nil {
		return roomdomain.Room{}, err
	}

	room, err := r.loadRoomByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roomdomain.Room{}, roomdomain.ErrRoomNotFound
		}
		return roomdomain.Room{}, err
	}
	return room, nil
}

func (r *Repository) Create(ctx context.Context, input roomusecase.CreateInput) (roomdomain.Room, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return roomdomain.Room{}, err
		}
	}
	if err := r.ensureRoomTypeInDormitory(ctx, input.RoomTypeID, input.DormitoryID); err != nil {
		return roomdomain.Room{}, err
	}

	room := roomdomain.Room{
		ID:          uuid.New(),
		DormitoryID: input.DormitoryID,
		RoomTypeID:  input.RoomTypeID,
		RoomNumber:  input.RoomNumber,
		Floor:       input.Floor,
		Status:      input.Status,
		IsActive:    input.IsActive,
		CreatedBy:   input.CreatedBy,
		UpdatedBy:   input.CreatedBy,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO rooms (id, dormitory_id, room_type_id, room_number, floor, status, is_active, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`, room.ID, room.DormitoryID, room.RoomTypeID, room.RoomNumber, room.Floor, room.Status, room.IsActive, room.CreatedBy, room.UpdatedBy).
		Scan(&room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return roomdomain.Room{}, roomdomain.ErrRoomNumberExists
			}
			if pgErr.Code == "23503" {
				return roomdomain.Room{}, roomdomain.ErrDormitoryNotFound
			}
		}
		return roomdomain.Room{}, err
	}

	return r.loadRoomByID(ctx, room.ID)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input roomusecase.UpdateInput) (roomdomain.Room, error) {
	if err := r.ensureRoomAccess(ctx, id, requesterID); err != nil {
		return roomdomain.Room{}, err
	}

	if input.RoomTypeID != nil {
		var dormitoryID uuid.UUID
		if err := r.db.QueryRow(ctx, `SELECT dormitory_id FROM rooms WHERE id = $1`, id).Scan(&dormitoryID); err != nil {
			return roomdomain.Room{}, err
		}
		if err := r.ensureRoomTypeInDormitory(ctx, *input.RoomTypeID, dormitoryID); err != nil {
			return roomdomain.Room{}, err
		}
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.RoomTypeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("room_type_id = $%d", argIdx))
		args = append(args, *input.RoomTypeID)
		argIdx++
	}
	if input.RoomNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("room_number = $%d", argIdx))
		args = append(args, *input.RoomNumber)
		argIdx++
	}
	if input.Floor != nil {
		setClauses = append(setClauses, fmt.Sprintf("floor = $%d", argIdx))
		args = append(args, *input.Floor)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
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
		query := fmt.Sprintf("UPDATE rooms SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return roomdomain.Room{}, roomdomain.ErrRoomNumberExists
			}
			return roomdomain.Room{}, err
		}
	}

	return r.loadRoomByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureRoomAccess(ctx, id, requesterID); err != nil {
		return err
	}

	contractCount, err := r.CountContracts(ctx, id)
	if err != nil {
		return err
	}
	if contractCount > 0 {
		return roomdomain.ErrRoomHasContracts
	}

	result, err := r.db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return roomdomain.ErrRoomNotFound
	}

	return nil
}

func (r *Repository) CountContracts(ctx context.Context, id uuid.UUID) (int64, error) {
	var contractCount int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contracts WHERE room_id = $1`, id).Scan(&contractCount)
	return contractCount, err
}

func (r *Repository) loadRoomByID(ctx context.Context, id uuid.UUID) (roomdomain.Room, error) {
	var room roomdomain.Room
	err := r.db.QueryRow(ctx, `
		SELECT rm.id, rm.dormitory_id, d.name, rm.room_type_id, rt.name, rt.price, rm.room_number, rm.floor, rm.status, rm.is_active, rm.created_by, rm.updated_by, rm.created_at, rm.updated_at
		FROM rooms rm
		JOIN dormitories d ON d.id = rm.dormitory_id
		JOIN room_types rt ON rt.id = rm.room_type_id
		WHERE rm.id = $1
	`, id).Scan(
		&room.ID,
		&room.DormitoryID,
		&room.DormitoryName,
		&room.RoomTypeID,
		&room.RoomTypeName,
		&room.RoomTypePrice,
		&room.RoomNumber,
		&room.Floor,
		&room.Status,
		&room.IsActive,
		&room.CreatedBy,
		&room.UpdatedBy,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err != nil {
		return roomdomain.Room{}, err
	}
	return room, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages rooms in every dormitory), along with their role
// ID so callers can also check role-level dormitory grants.
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
// manage rooms under it (unrestricted, individually assigned via
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
			return roomdomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureRoomTypeInDormitory confirms the room type exists and belongs to the
// given dormitory, so a room can't be created against a mismatched type.
func (r *Repository) ensureRoomTypeInDormitory(ctx context.Context, roomTypeID, dormitoryID uuid.UUID) error {
	var exists int
	err := r.db.QueryRow(ctx, `
		SELECT 1 FROM room_types WHERE id = $1 AND dormitory_id = $2
	`, roomTypeID, dormitoryID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roomdomain.ErrRoomTypeNotFound
		}
		return err
	}
	return nil
}

// ensureRoomAccess confirms the room exists and the requester may act on it,
// based on access to its parent dormitory. Both a missing room and a missing
// grant surface as ErrRoomNotFound so scoped-out callers can't distinguish
// the two.
func (r *Repository) ensureRoomAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM rooms rm
		WHERE rm.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return roomdomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}
