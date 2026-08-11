package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/features/announcement/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type AnnouncementRepo struct {
	db *database.DB
}

func NewAnnouncementRepo(db *database.DB) *AnnouncementRepo {
	return &AnnouncementRepo{db: db}
}

func scanAnnouncement(row pgx.Row) (*domain.Announcement, error) {
	a := &domain.Announcement{}
	err := row.Scan(
		&a.ID, &a.DormitoryID, &a.Title, &a.Content, &a.Type, &a.IsPinned,
		&a.PublishedAt, &a.ExpiredAt, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *AnnouncementRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.Announcement, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id, dormitory_id, title, content, type, is_pinned, published_at, expired_at, created_at, updated_at
		FROM announcements WHERE id = $1 AND dormitory_id = $2`, id, dormitoryID)
	a, err := scanAnnouncement(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("announcement not found: %w", coredomain.ErrNotFound)
	}
	return a, err
}

func (r *AnnouncementRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.Announcement, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, dormitory_id, title, content, type, is_pinned, published_at, expired_at, created_at, updated_at
		FROM announcements
		WHERE dormitory_id = $1
		ORDER BY is_pinned DESC, published_at DESC, created_at DESC
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Announcement
	for rows.Next() {
		a, err := scanAnnouncement(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	if list == nil {
		list = []*domain.Announcement{}
	}
	return list, rows.Err()
}

func (r *AnnouncementRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM announcements WHERE dormitory_id = $1`, dormitoryID).Scan(&total)
	return total, err
}

func (r *AnnouncementRepo) Create(ctx context.Context, a *domain.Announcement) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO announcements (id, dormitory_id, title, content, type, is_pinned, published_at, expired_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.DormitoryID, a.Title, a.Content, a.Type, a.IsPinned, a.PublishedAt, a.ExpiredAt)
	return err
}

func (r *AnnouncementRepo) Update(ctx context.Context, a *domain.Announcement) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE announcements
		SET title=$3, content=$4, type=$5, is_pinned=$6, published_at=$7, expired_at=$8, updated_at=NOW()
		WHERE id=$1 AND dormitory_id=$2`,
		a.ID, a.DormitoryID, a.Title, a.Content, a.Type, a.IsPinned, a.PublishedAt, a.ExpiredAt)
	return err
}

func (r *AnnouncementRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM announcements WHERE id = $1 AND dormitory_id = $2`, id, dormitoryID)
	return err
}
