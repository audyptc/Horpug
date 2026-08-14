package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	repairrequestdomain "apihorpug/internal/features/repairrequest/domain"
	repairrequestusecase "apihorpug/internal/features/repairrequest/usecase"

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

const selectRepairRequestColumns = `
	rr.id, rr.room_id, rm.room_number, rm.dormitory_id, d.name, rr.tenant_id,
	t.first_name || ' ' || t.last_name, rr.category, rr.description, rr.status, rr.reported_date,
	rr.created_by, rr.updated_by, rr.created_at, rr.updated_at
`

const repairRequestFromJoins = `
	FROM repair_requests rr
	JOIN rooms rm ON rm.id = rr.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
	LEFT JOIN tenants t ON t.id = rr.tenant_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters repairrequestusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`rr.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`rr.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf(`rr.category = $%d`, *argIdx))
		*args = append(*args, *filters.Category)
		*argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf(`rr.status = $%d`, *argIdx))
		*args = append(*args, *filters.Status)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters repairrequestusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + repairRequestFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters repairrequestusecase.ListFilters, limit, offset int) ([]repairrequestdomain.RepairRequest, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectRepairRequestColumns + repairRequestFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY rr.reported_date DESC, rr.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanRepairRequests(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (repairrequestdomain.RepairRequest, error) {
	if err := r.ensureRepairRequestAccess(ctx, id, requesterID); err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}

	request, err := r.loadRepairRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRepairRequestNotFound
		}
		return repairrequestdomain.RepairRequest{}, err
	}

	return request, nil
}

func (r *Repository) Create(ctx context.Context, input repairrequestusecase.CreateInput) (repairrequestdomain.RepairRequest, error) {
	if input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, input.RoomID, *input.CreatedBy); err != nil {
			return repairrequestdomain.RepairRequest{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO repair_requests (id, room_id, tenant_id, category, description, status, reported_date, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
	`, id, input.RoomID, input.TenantID, input.Category, input.Description, input.Status, input.ReportedDate, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			if pgErr.ConstraintName == "repair_requests_tenant_fkey" {
				return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrTenantNotFound
			}
			return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrRoomNotFound
		}
		return repairrequestdomain.RepairRequest{}, err
	}

	return r.loadRepairRequestByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input repairrequestusecase.UpdateInput) (repairrequestdomain.RepairRequest, error) {
	if err := r.ensureRepairRequestAccess(ctx, id, requesterID); err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.TenantID != nil {
		setClauses = append(setClauses, fmt.Sprintf("tenant_id = $%d", argIdx))
		args = append(args, *input.TenantID)
		argIdx++
	}
	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *input.Category)
		argIdx++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *input.Description)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.ReportedDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("reported_date = $%d", argIdx))
		args = append(args, *input.ReportedDate)
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
		query := fmt.Sprintf("UPDATE repair_requests SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "repair_requests_tenant_fkey" {
				return repairrequestdomain.RepairRequest{}, repairrequestdomain.ErrTenantNotFound
			}
			return repairrequestdomain.RepairRequest{}, err
		}
	}

	return r.loadRepairRequestByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureRepairRequestAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM repair_requests WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return repairrequestdomain.ErrRepairRequestNotFound
	}

	return nil
}

func (r *Repository) loadRepairRequestByID(ctx context.Context, id uuid.UUID) (repairrequestdomain.RepairRequest, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectRepairRequestColumns+repairRequestFromJoins+` WHERE rr.id = $1`, id)
	return scanRepairRequest(row)
}

func scanRepairRequest(row pgx.Row) (repairrequestdomain.RepairRequest, error) {
	var request repairrequestdomain.RepairRequest
	if err := row.Scan(
		&request.ID,
		&request.RoomID,
		&request.RoomNumber,
		&request.DormitoryID,
		&request.DormitoryName,
		&request.TenantID,
		&request.TenantName,
		&request.Category,
		&request.Description,
		&request.Status,
		&request.ReportedDate,
		&request.CreatedBy,
		&request.UpdatedBy,
		&request.CreatedAt,
		&request.UpdatedAt,
	); err != nil {
		return repairrequestdomain.RepairRequest{}, err
	}
	return request, nil
}

func scanRepairRequests(rows pgx.Rows) ([]repairrequestdomain.RepairRequest, error) {
	requests := make([]repairrequestdomain.RepairRequest, 0)
	for rows.Next() {
		request, err := scanRepairRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return requests, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages repair requests in every dormitory), along with
// their role ID so callers can also check role-level dormitory grants.
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

// ensureRoomAccess confirms the room exists and the requester may record
// repair requests against it, based on access to its parent dormitory.
func (r *Repository) ensureRoomAccess(ctx context.Context, roomID, requesterID uuid.UUID) error {
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
	`, roomID, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repairrequestdomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}

// ensureRepairRequestAccess confirms the repair request exists and the
// requester may act on it, based on access to the dormitory of its room.
// Both a missing request and a missing grant surface as
// ErrRepairRequestNotFound so scoped-out callers can't distinguish the two.
func (r *Repository) ensureRepairRequestAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM repair_requests rr
		JOIN rooms rm ON rm.id = rr.room_id
		WHERE rr.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repairrequestdomain.ErrRepairRequestNotFound
		}
		return err
	}
	return nil
}
