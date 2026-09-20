package postgres

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	announcementdomain "apihorpug/internal/features/announcement/domain"
	announcementusecase "apihorpug/internal/features/announcement/usecase"
	"apihorpug/internal/platform/sqlutil"

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

// selectAnnouncementColumns embeds the requester as a literal for is_read. It is
// a uuid.UUID rendered in canonical form, so it can't carry SQL, and it saves
// threading one more positional argument through every query that selects rows.
func selectAnnouncementColumns(requesterID uuid.UUID) string {
	return fmt.Sprintf(`
	a.id, a.dormitory_id, d.name, a.title, a.content, a.category, a.is_pinned, a.is_published, a.published_date,
	EXISTS (SELECT 1 FROM announcement_reads ar WHERE ar.announcement_id = a.id AND ar.user_id = '%s'),
	a.created_by, a.updated_by, a.created_at, a.updated_at
`, requesterID)
}

const announcementFromJoins = `
	FROM announcements a
	JOIN dormitories d ON d.id = a.dormitory_id
`

// visibleToTenants is what someone who can't manage announcements may see:
// only published ones whose date has arrived. The date is judged in Thailand
// time, as the rest of the system is, so a notice dated "today" appears from
// local midnight rather than seven hours later.
const visibleToTenants = `(a.is_published = TRUE AND a.published_date <= (NOW() AT TIME ZONE 'Asia/Bangkok')::date)`

// viewer is how the requester's role decides what they can see.
type viewer struct {
	// full roles are exempt from per-dormitory scoping.
	full   bool
	roleID uuid.UUID
	// canManage roles (any of create/update/delete on the announcements menu)
	// also see drafts and announcements dated in the future.
	canManage bool
}

func (r *Repository) buildScope(v viewer, requesterID uuid.UUID, filters announcementusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !v.full {
		conditions = append(conditions, fmt.Sprintf(`a.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, v.roleID)
		*argIdx += 2
	}
	if !v.canManage {
		conditions = append(conditions, visibleToTenants)
	}
	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf(`a.category = $%d`, *argIdx))
		*args = append(*args, filters.Category)
		*argIdx++
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

	if filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf(`(d.name ILIKE $%d OR a.title ILIKE $%d OR a.content ILIKE $%d)`, *argIdx, *argIdx, *argIdx))
		*args = append(*args, sqlutil.ContainsPattern(filters.Search))
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
		value := filters.Columns[key]
		var clause string
		switch key {
		case "dormitory_name":
			clause = fmt.Sprintf(`d.name ILIKE $%d`, *argIdx)
		case "title":
			clause = fmt.Sprintf(`a.title ILIKE $%d`, *argIdx)
		default:
			continue
		}
		conditions = append(conditions, clause)
		*args = append(*args, sqlutil.ContainsPattern(value))
		*argIdx++
	}

	return conditions
}

// listOrderBy resolves the sort key through the usecase whitelist; anything
// unrecognised falls back to the default rather than reaching SQL. a.id
// breaks ties so paging over equal values can't repeat or skip a row.
func listOrderBy(filters announcementusecase.ListFilters) string {
	sortKey := filters.SortKey
	if _, ok := announcementusecase.SortColumns[sortKey]; !ok {
		sortKey = announcementusecase.DefaultSortKey
	}

	direction := "ASC"
	if filters.SortDesc {
		direction = "DESC"
	}

	switch sortKey {
	case "dormitory_name":
		return fmt.Sprintf(" ORDER BY d.name %s, a.id ASC", direction)
	case "title":
		return fmt.Sprintf(" ORDER BY a.title %s, a.id ASC", direction)
	case "category":
		return fmt.Sprintf(" ORDER BY a.category %s, a.id ASC", direction)
	case "is_published":
		return fmt.Sprintf(" ORDER BY a.is_published %s, a.id ASC", direction)
	default: // published_date; pinned notices stay on top whichever way dates run
		return fmt.Sprintf(" ORDER BY a.is_pinned DESC, a.published_date %s, a.id ASC", direction)
	}
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters announcementusecase.ListFilters) (int64, error) {
	v, err := r.viewerFor(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(v, requesterID, filters, &argIdx, &args)

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
	v, err := r.viewerFor(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(v, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectAnnouncementColumns(requesterID) + announcementFromJoins
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

	return scanAnnouncements(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error) {
	if err := r.ensureAnnouncementAccess(ctx, id, requesterID); err != nil {
		return announcementdomain.Announcement{}, err
	}

	announcement, err := r.loadAnnouncementByID(ctx, id, requesterID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return announcementdomain.Announcement{}, announcementdomain.ErrAnnouncementNotFound
		}
		return announcementdomain.Announcement{}, err
	}

	return announcement, nil
}

// Summary counts the announcements the requester can see but hasn't opened.
// Drafts and not-yet-dated ones never count, even for managers: they aren't
// news to anyone yet.
func (r *Repository) Summary(ctx context.Context, requesterID uuid.UUID) (announcementdomain.Summary, error) {
	v, err := r.viewerFor(ctx, requesterID)
	if err != nil {
		return announcementdomain.Summary{}, err
	}

	argIdx := 1
	args := make([]any, 0)
	audience := viewer{full: v.full, roleID: v.roleID, canManage: false}
	conditions := r.buildScope(audience, requesterID, announcementusecase.ListFilters{}, &argIdx, &args)
	conditions = append(conditions, fmt.Sprintf(
		`NOT EXISTS (SELECT 1 FROM announcement_reads ar WHERE ar.announcement_id = a.id AND ar.user_id = $%d)`, argIdx))
	args = append(args, requesterID)

	query := `SELECT COUNT(*) ` + announcementFromJoins + ` WHERE ` + strings.Join(conditions, " AND ")

	summary := announcementdomain.Summary{CanManage: v.canManage}
	if err := r.db.QueryRow(ctx, query, args...).Scan(&summary.UnreadCount); err != nil {
		return announcementdomain.Summary{}, err
	}
	return summary, nil
}

func (r *Repository) MarkRead(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureAnnouncementAccess(ctx, id, requesterID); err != nil {
		return err
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO announcement_reads (announcement_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (announcement_id, user_id) DO NOTHING
	`, id, requesterID)
	return err
}

func (r *Repository) Create(ctx context.Context, input announcementusecase.CreateInput) (announcementdomain.Announcement, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return announcementdomain.Announcement{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO announcements (id, dormitory_id, title, content, category, is_pinned, is_published, published_date, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
	`, id, input.DormitoryID, input.Title, input.Content, input.Category, input.IsPinned, *input.IsPublished, input.PublishedDate, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return announcementdomain.Announcement{}, announcementdomain.ErrDormitoryNotFound
		}
		return announcementdomain.Announcement{}, err
	}

	return r.loadAnnouncementByID(ctx, id, requesterOrNil(input.CreatedBy))
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
	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *input.Category)
		argIdx++
	}
	if input.IsPinned != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_pinned = $%d", argIdx))
		args = append(args, *input.IsPinned)
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

	return r.loadAnnouncementByID(ctx, id, requesterID)
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

func (r *Repository) loadAnnouncementByID(ctx context.Context, id, requesterID uuid.UUID) (announcementdomain.Announcement, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectAnnouncementColumns(requesterID)+announcementFromJoins+` WHERE a.id = $1`, id)
	return scanAnnouncement(row)
}

// requesterOrNil lets a create with no recorded author still read back its row;
// uuid.Nil matches no reader, so is_read comes out false.
func requesterOrNil(id *uuid.UUID) uuid.UUID {
	if id == nil {
		return uuid.Nil
	}
	return *id
}

func scanAnnouncement(row pgx.Row) (announcementdomain.Announcement, error) {
	var announcement announcementdomain.Announcement
	if err := row.Scan(
		&announcement.ID,
		&announcement.DormitoryID,
		&announcement.DormitoryName,
		&announcement.Title,
		&announcement.Content,
		&announcement.Category,
		&announcement.IsPinned,
		&announcement.IsPublished,
		&announcement.PublishedDate,
		&announcement.IsRead,
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

// viewerFor resolves how the user's role scopes what they can see: whether it
// is exempt from per-dormitory scoping, its role ID (for role-level dormitory
// grants), and whether it can manage announcements at all. Managing means
// holding any of create/update/delete on the announcements menu; a role that
// can only read is treated as a tenant and never sees drafts.
func (r *Repository) viewerFor(ctx context.Context, userID uuid.UUID) (viewer, error) {
	var v viewer
	err := r.db.QueryRow(ctx, `
		SELECT r.full_dormitory_access, r.id, EXISTS (
			SELECT 1
			FROM role_menu_permissions rmp
			JOIN menus m ON m.id = rmp.menu_id
			JOIN permissions p ON p.id = rmp.permission_id
			WHERE rmp.role_id = r.id AND m.path = '/announcements'
			AND p.name IN ('create', 'update', 'delete')
		)
		FROM users u
		JOIN roles r ON r.id = u.role_id
		WHERE u.id = $1
	`, userID).Scan(&v.full, &v.roleID, &v.canManage)
	if err != nil {
		return viewer{}, err
	}
	return v, nil
}

// ensureDormitoryAccess confirms the dormitory exists and the requester may
// post announcements under it (unrestricted, individually assigned via
// user_dormitories, or granted through their role via role_dormitories).
func (r *Repository) ensureDormitoryAccess(ctx context.Context, dormitoryID, requesterID uuid.UUID) error {
	v, err := r.viewerFor(ctx, requesterID)
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
	`, dormitoryID, v.full, requesterID, v.roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return announcementdomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureAnnouncementAccess confirms the announcement exists and the
// requester may act on it, based on access to its parent dormitory. Someone
// who can't manage announcements also can't reach a draft or one dated in the
// future by id. A missing announcement, a missing grant and a hidden draft all
// surface as ErrAnnouncementNotFound so callers can't tell them apart.
func (r *Repository) ensureAnnouncementAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	v, err := r.viewerFor(ctx, requesterID)
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
		AND ($5 OR `+visibleToTenants+`)
	`, id, v.full, requesterID, v.roleID, v.canManage).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return announcementdomain.ErrAnnouncementNotFound
		}
		return err
	}
	return nil
}
