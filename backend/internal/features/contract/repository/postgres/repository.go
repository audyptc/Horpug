package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	contractdomain "apihorpug/internal/features/contract/domain"
	contractusecase "apihorpug/internal/features/contract/usecase"

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

const selectContractColumns = `
	c.id, c.tenant_id, t.first_name, t.last_name,
	c.room_id, rm.room_number, rm.dormitory_id, d.name,
	c.start_date, c.end_date, c.rent_price, c.deposit, c.num_occupants, c.status, c.note,
	c.created_by, c.updated_by, c.created_at, c.updated_at
`

const contractFromJoins = `
	FROM contracts c
	JOIN tenants t ON t.id = c.tenant_id
	JOIN rooms rm ON rm.id = c.room_id
	JOIN dormitories d ON d.id = rm.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters contractusecase.ListFilters, argIdx *int, args *[]any) []string {
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
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`c.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`rm.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf(`c.status = $%d`, *argIdx))
		*args = append(*args, *filters.Status)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters contractusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + contractFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters contractusecase.ListFilters, limit, offset int) ([]contractdomain.Contract, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectContractColumns + contractFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY c.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanContracts(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (contractdomain.Contract, error) {
	if err := r.ensureContractAccess(ctx, id, requesterID); err != nil {
		return contractdomain.Contract{}, err
	}

	contract, err := r.loadContractByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contractdomain.Contract{}, contractdomain.ErrContractNotFound
		}
		return contractdomain.Contract{}, err
	}
	return contract, nil
}

func (r *Repository) Create(ctx context.Context, input contractusecase.CreateInput) (contractdomain.Contract, error) {
	if input.CreatedBy != nil {
		if err := r.ensureRoomAccess(ctx, input.RoomID, *input.CreatedBy); err != nil {
			return contractdomain.Contract{}, err
		}
	}

	id := uuid.New()
	err := r.db.QueryRow(ctx, `
		INSERT INTO contracts (id, tenant_id, room_id, start_date, end_date, rent_price, deposit, num_occupants, status, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`, id, input.TenantID, input.RoomID, input.StartDate, input.EndDate, input.RentPrice, input.Deposit, input.NumOccupants,
		contractdomain.ContractStatusActive, input.Note, input.CreatedBy, input.CreatedBy).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return contractdomain.Contract{}, contractdomain.ErrRoomHasActiveContract
			}
			if pgErr.Code == "23503" {
				if pgErr.ConstraintName == "contracts_tenant_fkey" {
					return contractdomain.Contract{}, contractdomain.ErrTenantNotFound
				}
				return contractdomain.Contract{}, contractdomain.ErrRoomNotFound
			}
		}
		return contractdomain.Contract{}, err
	}

	return r.loadContractByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input contractusecase.UpdateInput) (contractdomain.Contract, error) {
	if err := r.ensureContractAccess(ctx, id, requesterID); err != nil {
		return contractdomain.Contract{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.EndDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("end_date = $%d", argIdx))
		args = append(args, *input.EndDate)
		argIdx++
	}
	if input.RentPrice != nil {
		setClauses = append(setClauses, fmt.Sprintf("rent_price = $%d", argIdx))
		args = append(args, *input.RentPrice)
		argIdx++
	}
	if input.Deposit != nil {
		setClauses = append(setClauses, fmt.Sprintf("deposit = $%d", argIdx))
		args = append(args, *input.Deposit)
		argIdx++
	}
	if input.NumOccupants != nil {
		setClauses = append(setClauses, fmt.Sprintf("num_occupants = $%d", argIdx))
		args = append(args, *input.NumOccupants)
		argIdx++
	}
	if input.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *input.Status)
		argIdx++
	}
	if input.Note != nil {
		setClauses = append(setClauses, fmt.Sprintf("note = $%d", argIdx))
		args = append(args, *input.Note)
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
		query := fmt.Sprintf("UPDATE contracts SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return contractdomain.Contract{}, contractdomain.ErrRoomHasActiveContract
			}
			return contractdomain.Contract{}, err
		}
	}

	return r.loadContractByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureContractAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM contracts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return contractdomain.ErrContractNotFound
	}

	return nil
}

func (r *Repository) loadContractByID(ctx context.Context, id uuid.UUID) (contractdomain.Contract, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectContractColumns+contractFromJoins+` WHERE c.id = $1`, id)
	return scanContract(row)
}

func scanContract(row pgx.Row) (contractdomain.Contract, error) {
	var contract contractdomain.Contract
	var firstName, lastName string
	if err := row.Scan(
		&contract.ID,
		&contract.TenantID,
		&firstName,
		&lastName,
		&contract.RoomID,
		&contract.RoomNumber,
		&contract.DormitoryID,
		&contract.DormitoryName,
		&contract.StartDate,
		&contract.EndDate,
		&contract.RentPrice,
		&contract.Deposit,
		&contract.NumOccupants,
		&contract.Status,
		&contract.Note,
		&contract.CreatedBy,
		&contract.UpdatedBy,
		&contract.CreatedAt,
		&contract.UpdatedAt,
	); err != nil {
		return contractdomain.Contract{}, err
	}
	contract.TenantName = strings.TrimSpace(firstName + " " + lastName)
	return contract, nil
}

func scanContracts(rows pgx.Rows) ([]contractdomain.Contract, error) {
	contracts := make([]contractdomain.Contract, 0)
	for rows.Next() {
		contract, err := scanContract(rows)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return contracts, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages contracts in every dormitory), along with their
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

// ensureRoomAccess confirms the room exists and the requester may create
// contracts against it, based on access to its parent dormitory.
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
			return contractdomain.ErrRoomNotFound
		}
		return err
	}
	return nil
}

// ensureContractAccess confirms the contract exists and the requester may act
// on it, based on access to the dormitory of its room. Both a missing
// contract and a missing grant surface as ErrContractNotFound so scoped-out
// callers can't distinguish the two.
func (r *Repository) ensureContractAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM contracts c
		JOIN rooms rm ON rm.id = c.room_id
		WHERE c.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = rm.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = rm.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return contractdomain.ErrContractNotFound
		}
		return err
	}
	return nil
}
