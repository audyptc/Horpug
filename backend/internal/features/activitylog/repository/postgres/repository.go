package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
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

// buildListConditions builds the WHERE conditions shared by Count and List.
// Both must apply exactly the same conditions, otherwise the reported total
// disagrees with the rows returned and the client paginates over a page
// count that doesn't exist.
func buildListConditions(filter activitylogusecase.ListFilter, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("al.user_id = $%d", *argIdx))
		*args = append(*args, *filter.UserID)
		*argIdx++
	}
	if filter.EntityID != nil {
		conditions = append(conditions, fmt.Sprintf("al.entity_id = $%d", *argIdx))
		*args = append(*args, *filter.EntityID)
		*argIdx++
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf("al.created_at >= $%d", *argIdx))
		*args = append(*args, *filter.DateFrom)
		*argIdx++
	}
	if filter.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf("al.created_at < $%d", *argIdx))
		*args = append(*args, *filter.DateTo)
		*argIdx++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			`(COALESCE(u.username, '') ILIKE $%d OR al.action ILIKE $%d OR al.entity_type ILIKE $%d OR al.description ILIKE $%d OR al.ip_address ILIKE $%d)`,
			*argIdx, *argIdx, *argIdx, *argIdx, *argIdx,
		))
		*args = append(*args, "%"+filter.Search+"%")
		*argIdx++
	}

	// Sorted so the generated SQL is stable for a given set of filters rather
	// than varying with Go's randomised map iteration order.
	keys := make([]string, 0, len(filter.Columns))
	for key := range filter.Columns {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		column, ok := activitylogusecase.FilterColumns[key]
		if !ok {
			continue
		}
		conditions = append(conditions, fmt.Sprintf("%s ILIKE $%d", column, *argIdx))
		*args = append(*args, "%"+filter.Columns[key]+"%")
		*argIdx++
	}

	return conditions
}

// listOrderBy resolves the sort key through the usecase whitelist; anything
// unrecognised falls back to the default rather than reaching SQL. al.id
// breaks ties so paging over equal values can't repeat or skip a row.
func listOrderBy(filter activitylogusecase.ListFilter) string {
	column, ok := activitylogusecase.SortColumns[filter.SortKey]
	if !ok {
		column = activitylogusecase.SortColumns[activitylogusecase.DefaultSortKey]
	}

	direction := "ASC"
	if filter.SortDesc {
		direction = "DESC"
	}

	return fmt.Sprintf(" ORDER BY %s %s, al.id ASC", column, direction)
}

func (r *Repository) Count(ctx context.Context, filter activitylogusecase.ListFilter) (int64, error) {
	argIdx := 1
	args := make([]any, 0)
	conditions := buildListConditions(filter, &argIdx, &args)

	query := `SELECT COUNT(*) FROM activity_logs al LEFT JOIN users u ON u.id = al.user_id`
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
	argIdx := 1
	args := make([]any, 0)
	conditions := buildListConditions(filter, &argIdx, &args)

	query := fmt.Sprintf(`
		SELECT %s
		FROM activity_logs al
		LEFT JOIN users u ON u.id = al.user_id
	`, selectColumns)
	if len(conditions) > 0 {
		query += "WHERE " + strings.Join(conditions, " AND ") + " "
	}
	query += listOrderBy(filter) + fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
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
