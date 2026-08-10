package repository

import (
	"context"
	"fmt"
	"strings"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/feature/activitylog/domain"
)

type ActivityLogRepo struct {
	db *database.DB
}

func NewActivityLogRepo(db *database.DB) *ActivityLogRepo {
	return &ActivityLogRepo{db: db}
}

func (r *ActivityLogRepo) Create(ctx context.Context, log *domain.ActivityLog) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO activity_logs (id, actor_id, dormitory_id, action, entity_type, entity_id, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.ID, log.ActorID, log.DormitoryID, log.Action, log.EntityType, log.EntityID, log.NewValue, log.CreatedAt)
	return err
}

func (r *ActivityLogRepo) buildFilter(filter domain.ActivityLogFilter) (string, []interface{}) {
	var where []string
	var args []interface{}
	n := 1

	if filter.DormitoryID != "" {
		where = append(where, fmt.Sprintf("al.dormitory_id = $%d", n))
		args = append(args, filter.DormitoryID)
		n++
	}
	if filter.EntityType != "" {
		where = append(where, fmt.Sprintf("al.entity_type = $%d", n))
		args = append(args, filter.EntityType)
		n++
	}
	if filter.ActorID != "" {
		where = append(where, fmt.Sprintf("al.actor_id = $%d", n))
		args = append(args, filter.ActorID)
		n++
	}
	if filter.From != nil {
		where = append(where, fmt.Sprintf("al.created_at >= $%d", n))
		args = append(args, filter.From)
		n++
	}
	if filter.To != nil {
		where = append(where, fmt.Sprintf("al.created_at <= $%d", n))
		args = append(args, filter.To)
	}

	clause := ""
	if len(where) > 0 {
		clause = "WHERE " + strings.Join(where, " AND ")
	}
	return clause, args
}

func (r *ActivityLogRepo) List(ctx context.Context, filter domain.ActivityLogFilter, limit, offset int) ([]*domain.ActivityLog, error) {
	where, args := r.buildFilter(filter)
	n := len(args) + 1

	query := fmt.Sprintf(`
		SELECT al.id, al.actor_id, COALESCE(u.full_name, ''), al.dormitory_id, al.action, al.entity_type, al.entity_id, al.new_value, al.created_at
		FROM activity_logs al
		LEFT JOIN users u ON u.id = al.actor_id
		%s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d`, where, n, n+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.ActivityLog
	for rows.Next() {
		l := &domain.ActivityLog{}
		if err := rows.Scan(&l.ID, &l.ActorID, &l.ActorName, &l.DormitoryID, &l.Action, &l.EntityType, &l.EntityID, &l.NewValue, &l.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	if list == nil {
		list = []*domain.ActivityLog{}
	}
	return list, rows.Err()
}

func (r *ActivityLogRepo) Count(ctx context.Context, filter domain.ActivityLogFilter) (int, error) {
	where, args := r.buildFilter(filter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM activity_logs al %s`, where)
	var total int
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(&total)
	return total, err
}
