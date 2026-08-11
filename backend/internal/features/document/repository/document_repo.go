package repository

import (
	"context"
	"fmt"

	"apigofiberhorpug/internal/features/document/domain"
	"apigofiberhorpug/internal/platform/database"
	coredomain "apigofiberhorpug/internal/shared/domain"

	"github.com/jackc/pgx/v5"
)

type DocumentRepo struct {
	db *database.DB
}

func NewDocumentRepo(db *database.DB) *DocumentRepo {
	return &DocumentRepo{db: db}
}

func scanDocumentDetail(row pgx.Row) (*domain.DocumentDetail, error) {
	d := &domain.DocumentDetail{}
	err := row.Scan(
		&d.ID, &d.DormitoryID, &d.Title, &d.Category, &d.TenantID,
		&d.FileURL, &d.IssueDate, &d.ExpiryDate,
		&d.Note, &d.CreatedAt, &d.UpdatedAt,
		&d.TenantFirstName, &d.TenantLastName,
	)
	return d, err
}

const selectDocumentDetail = `
	SELECT
		d.id, d.dormitory_id, d.title, d.category, d.tenant_id,
		d.file_url, d.issue_date, d.expiry_date,
		d.note, d.created_at, d.updated_at,
		COALESCE(t.first_name, '') AS tenant_first_name,
		COALESCE(t.last_name, '') AS tenant_last_name
	FROM documents d
	LEFT JOIN tenants t ON t.id = d.tenant_id`

func (r *DocumentRepo) FindByID(ctx context.Context, dormitoryID, id string) (*domain.DocumentDetail, error) {
	row := r.db.Pool.QueryRow(ctx, selectDocumentDetail+` WHERE d.id = $1 AND d.dormitory_id = $2`, id, dormitoryID)
	d, err := scanDocumentDetail(row)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("document not found: %w", coredomain.ErrNotFound)
	}
	return d, err
}

func (r *DocumentRepo) List(ctx context.Context, dormitoryID string, limit, offset int) ([]*domain.DocumentDetail, error) {
	rows, err := r.db.Pool.Query(ctx, selectDocumentDetail+`
		WHERE d.dormitory_id = $1
		ORDER BY d.created_at DESC
		LIMIT $2 OFFSET $3`, dormitoryID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.DocumentDetail
	for rows.Next() {
		d, err := scanDocumentDetail(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	if list == nil {
		list = []*domain.DocumentDetail{}
	}
	return list, rows.Err()
}

func (r *DocumentRepo) Count(ctx context.Context, dormitoryID string) (int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM documents WHERE dormitory_id = $1`, dormitoryID).Scan(&total)
	return total, err
}

func (r *DocumentRepo) Create(ctx context.Context, d *domain.Document) error {
	_, err := r.db.Pool.Exec(ctx, `
		INSERT INTO documents (id, dormitory_id, title, category, tenant_id, file_url, issue_date, expiry_date, note)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		d.ID, d.DormitoryID, d.Title, d.Category, d.TenantID,
		d.FileURL, d.IssueDate, d.ExpiryDate, d.Note)
	return err
}

func (r *DocumentRepo) Update(ctx context.Context, d *domain.Document) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE documents
		SET title=$3, category=$4, tenant_id=$5, file_url=$6,
		    issue_date=$7, expiry_date=$8, note=$9, updated_at=NOW()
		WHERE id=$1 AND dormitory_id=$2`,
		d.ID, d.DormitoryID, d.Title, d.Category, d.TenantID,
		d.FileURL, d.IssueDate, d.ExpiryDate, d.Note)
	return err
}

func (r *DocumentRepo) Delete(ctx context.Context, dormitoryID, id string) error {
	_, err := r.db.Pool.Exec(ctx, `DELETE FROM documents WHERE id = $1 AND dormitory_id = $2`, id, dormitoryID)
	return err
}
