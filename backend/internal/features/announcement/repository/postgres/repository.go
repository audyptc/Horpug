package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	announcementdomain "apihorpug/internal/features/announcement/domain"
	announcementusecase "apihorpug/internal/features/announcement/usecase"

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

const selectAnnouncementColumns = `
	a.id, a.dormitory_id, d.name, a.title, a.content, a.is_published, a.published_date,
	a.created_by, a.updated_by, a.created_at, a.updated_at
`

const announcementFromJoins = `
	FROM announcements a
	JOIN dormitories d ON d.id = a.dormitory_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters announcementusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`a.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`a.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.IsPublished != nil {
		conditions = append(conditions, fmt.Sprintf(`a.is_published = $%d`, *argIdx))
		*args = append(*args, *filters.IsPublished)
		*argIdx++
	}
	if filters.DateFrom != nil {
		conditions = append(conditions, fmt.Sprintf(`a.published_date >= $%d`, *argIdx))
		*args = append(*args, *filters.DateFrom)
		*argIdx++
	}
	if filters.DateTo != nil {
		conditions = append(conditions, fmt.Sprintf(`a.published_date <= $%d`, *argIdx))
		*args = append(*args, *filters.DateTo)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters announcementusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + announcementFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters announcementusecase.ListFilters, limit, offset int) ([]announcementdomain.Announcement, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectAnnouncementColumns + announcementFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY a.published_date DESC, a.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAnnouncements(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error) {
	if err := r.ensureAnnouncementAccess(ctx, id, requesterID); err != nil {
		return announcementdomain.Announcement{}, err
	}

	announcement, err := r.loadAnnouncementByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return announcementdomain.Announcement{}, announcementdomain.ErrAnnouncementNotFound
		}
		return announcementdomain.Announcement{}, err
	}

	return announcement, nil
}

func (r *Repository) Create(ctx context.Context, input announcementusecase.CreateInput) (announcementdomain.Announcement, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return announcementdomain.Announcement{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO announcements (id, dormitory_id, title, content, is_published, published_date, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, input.DormitoryID, input.Title, input.Content, *input.IsPublished, input.PublishedDate, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return announcementdomain.Announcement{}, announcementdomain.ErrDormitoryNotFound
		}
		return announcementdomain.Announcement{}, err
	}

	return r.loadAnnouncementByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input announcementusecase.UpdateInput) (announcementdomain.Announcement, error) {
	if err := r.ensureAnnouncementAccess(ctx, id, requesterID); err != nil {
		return announcementdomain.Announcement{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIdx))
		args = append(args, *input.Title)
		argIdx++
	}
	if input.Content != nil {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argIdx))
		args = append(args, *input.Content)
		argIdx++
	}
	if input.IsPublished != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_published = $%d", argIdx))
		args = append(args, *input.IsPublished)
		argIdx++
	}
	if input.PublishedDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("published_date = $%d", argIdx))
		args = append(args, *input.PublishedDate)
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
		query := fmt.Sprintf("UPDATE announcements SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return announcementdomain.Announcement{}, err
		}
	}

	return r.loadAnnouncementByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureAnnouncementAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM announcements WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return announcementdomain.ErrAnnouncementNotFound
	}

	return nil
}

func (r *Repository) loadAnnouncementByID(ctx context.Context, id uuid.UUID) (announcementdomain.Announcement, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectAnnouncementColumns+announcementFromJoins+` WHERE a.id = $1`, id)
	return scanAnnouncement(row)
}

func scanAnnouncement(row pgx.Row) (announcementdomain.Announcement, error) {
	var announcement announcementdomain.Announcement
	if err := row.Scan(
		&announcement.ID,
		&announcement.DormitoryID,
		&announcement.DormitoryName,
		&announcement.Title,
		&announcement.Content,
		&announcement.IsPublished,
		&announcement.PublishedDate,
		&announcement.CreatedBy,
		&announcement.UpdatedBy,
		&announcement.CreatedAt,
		&announcement.UpdatedAt,
	); err != nil {
		return announcementdomain.Announcement{}, err
	}
	return announcement, nil
}

func scanAnnouncements(rows pgx.Rows) ([]announcementdomain.Announcement, error) {
	announcements := make([]announcementdomain.Announcement, 0)
	for rows.Next() {
		announcement, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		announcements = append(announcements, announcement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return announcements, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages announcements in every dormitory), along with
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

// ensureDormitoryAccess confirms the dormitory exists and the requester may
// post announcements under it (unrestricted, individually assigned via
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
			return announcementdomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureAnnouncementAccess confirms the announcement exists and the
// requester may act on it, based on access to its parent dormitory. Both a
// missing announcement and a missing grant surface as
// ErrAnnouncementNotFound so scoped-out callers can't distinguish the two.
func (r *Repository) ensureAnnouncementAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM announcements a
		WHERE a.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = a.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = a.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return announcementdomain.ErrAnnouncementNotFound
		}
		return err
	}
	return nil
}
