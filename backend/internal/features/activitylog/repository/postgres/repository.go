package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectColumns = `
	al.id,
	al.user_id,
	COALESCE(u.username, ''),
	al.action,
	al.entity_type,
	al.entity_id,
	al.description,
	al.ip_address,
	al.created_at
`

func scanActivityLog(row pgx.Row, log *activitylogdomain.ActivityLog) error {
	return row.Scan(
		&log.ID,
		&log.UserID,
		&log.Username,
		&log.Action,
		&log.EntityType,
		&log.EntityID,
		&log.Description,
		&log.IPAddress,
		&log.CreatedAt,
	)
}

func buildListConditions(filter activitylogusecase.ListFilter) ([]string, []any, int) {
	conditions := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("al.user_id = $%d", argIdx))
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("al.entity_type = $%d", argIdx))
		args = append(args, filter.EntityType)
		argIdx++
	}
	if filter.EntityID != nil {
		conditions = append(conditions, fmt.Sprintf("al.entity_id = $%d", argIdx))
		args = append(args, *filter.EntityID)
		argIdx++
	}

	return conditions, args, argIdx
}

func (r *Repository) Count(ctx context.Context, filter activitylogusecase.ListFilter) (int64, error) {
	conditions, args, _ := buildListConditions(filter)

	query := `SELECT COUNT(*) FROM activity_logs al`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, filter activitylogusecase.ListFilter) ([]activitylogdomain.ActivityLog, error) {
	conditions, args, argIdx := buildListConditions(filter)

	query := fmt.Sprintf(`
		SELECT %s
		FROM activity_logs al
		LEFT JOIN users u ON u.id = al.user_id
	`, selectColumns)
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + " "
	}
	query += fmt.Sprintf("ORDER BY al.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make([]activitylogdomain.ActivityLog, 0)
	for rows.Next() {
		var log activitylogdomain.ActivityLog
		if err := scanActivityLog(rows, &log); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (activitylogdomain.ActivityLog, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM activity_logs al
		LEFT JOIN users u ON u.id = al.user_id
		WHERE al.id = $1
	`, selectColumns)

	var log activitylogdomain.ActivityLog
	if err := scanActivityLog(r.db.QueryRow(ctx, query, id), &log); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return activitylogdomain.ActivityLog{}, activitylogdomain.ErrActivityLogNotFound
		}
		return activitylogdomain.ActivityLog{}, err
	}

	return log, nil
}

func (r *Repository) Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error) {
	log := activitylogdomain.ActivityLog{
		ID:          uuid.New(),
		UserID:      input.UserID,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		Description: input.Description,
		IPAddress:   input.IPAddress,
	}

	err := r.db.QueryRow(ctx, `
		INSERT INTO activity_logs (id, user_id, action, entity_type, entity_id, description, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`, log.ID, log.UserID, log.Action, log.EntityType, log.EntityID, log.Description, log.IPAddress).Scan(&log.CreatedAt)
	if err != nil {
		return activitylogdomain.ActivityLog{}, err
	}

	return r.GetByID(ctx, log.ID)
}
