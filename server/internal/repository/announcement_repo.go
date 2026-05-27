package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/domain"

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
		&a.ID, &a.Title, &a.Content, &a.Type, &a.IsPinned,
		&a.PublishedAt, &a.ExpiredAt, &a.CreatedAt, &a.UpdatedAt,
	)
	return a, err
}

func (r *AnnouncementRepo) FindByID(ctx context.Context, id string) (*domain.Announcement, error) {
	row := r.db.Pool.QueryRow(ctx, `
		SELECT id, title, content, type, is_pinned, published_at, expired_at, created_at, updated_at
		FROM announcements WHERE id = $1`, id)
	a, err := scanAnnouncement(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("announcement not found: %w", domain.ErrNotFound)
	}
	return a, err
}

func (r *AnnouncementRepo) List(ctx context.Context, limit, offset int) ([]*domain.Announcement, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT id, title, content, type, is_pinned, published_at, expired_at, created_at, updated_at
		FROM announcements
		ORDER BY is_pinned DESC, published_at DESC, created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
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

func (r *AnnouncementRepo) Count(ctx context.Context) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM announcements`).Scan(&total)
	return total, err
}

func (r *AnnouncementRepo) Create(ctx context.Context, a *domain.Announcement) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO announcements (id, title, content, type, is_pinned, published_at, expired_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.ID, a.Title, a.Content, a.Type, a.IsPinned, a.PublishedAt, a.ExpiredAt)
	return err
}

func (r *AnnouncementRepo) Update(ctx context.Context, a *domain.Announcement) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE announcements
		SET title=$2, content=$3, type=$4, is_pinned=$5, published_at=$6, expired_at=$7, updated_at=NOW()
		WHERE id=$1`,
		a.ID, a.Title, a.Content, a.Type, a.IsPinned, a.PublishedAt, a.ExpiredAt)
	return err
}

func (r *AnnouncementRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM announcements WHERE id = $1`, id)
	return err
}
