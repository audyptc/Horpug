package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	documentdomain "apihorpug/internal/features/document/domain"
	documentusecase "apihorpug/internal/features/document/usecase"

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

const selectDocumentColumns = `
	doc.id, doc.dormitory_id, d.name, doc.tenant_id, t.first_name || ' ' || t.last_name, doc.room_id, rm.room_number,
	doc.name, doc.category, doc.file_url, doc.uploaded_date, doc.note, doc.created_by, doc.updated_by, doc.created_at, doc.updated_at
`

const documentFromJoins = `
	FROM documents doc
	JOIN dormitories d ON d.id = doc.dormitory_id
	LEFT JOIN tenants t ON t.id = doc.tenant_id
	LEFT JOIN rooms rm ON rm.id = doc.room_id
`

func (r *Repository) buildScope(full bool, roleID, requesterID uuid.UUID, filters documentusecase.ListFilters, argIdx *int, args *[]any) []string {
	conditions := make([]string, 0)

	if !full {
		conditions = append(conditions, fmt.Sprintf(`doc.dormitory_id IN (
			SELECT dormitory_id FROM user_dormitories WHERE user_id = $%d
			UNION
			SELECT dormitory_id FROM role_dormitories WHERE role_id = $%d
		)`, *argIdx, *argIdx+1))
		*args = append(*args, requesterID, roleID)
		*argIdx += 2
	}
	if filters.DormitoryID != nil {
		conditions = append(conditions, fmt.Sprintf(`doc.dormitory_id = $%d`, *argIdx))
		*args = append(*args, *filters.DormitoryID)
		*argIdx++
	}
	if filters.TenantID != nil {
		conditions = append(conditions, fmt.Sprintf(`doc.tenant_id = $%d`, *argIdx))
		*args = append(*args, *filters.TenantID)
		*argIdx++
	}
	if filters.RoomID != nil {
		conditions = append(conditions, fmt.Sprintf(`doc.room_id = $%d`, *argIdx))
		*args = append(*args, *filters.RoomID)
		*argIdx++
	}
	if filters.Category != nil {
		conditions = append(conditions, fmt.Sprintf(`doc.category = $%d`, *argIdx))
		*args = append(*args, *filters.Category)
		*argIdx++
	}

	return conditions
}

func (r *Repository) Count(ctx context.Context, requesterID uuid.UUID, filters documentusecase.ListFilters) (int64, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return 0, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT COUNT(*) ` + documentFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int64
	if err := r.db.QueryRow(ctx, query, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) List(ctx context.Context, requesterID uuid.UUID, filters documentusecase.ListFilters, limit, offset int) ([]documentdomain.Document, error) {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return nil, err
	}

	argIdx := 1
	args := make([]any, 0)
	conditions := r.buildScope(full, roleID, requesterID, filters, &argIdx, &args)

	query := `SELECT ` + selectDocumentColumns + documentFromJoins
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += fmt.Sprintf(` ORDER BY doc.uploaded_date DESC, doc.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDocuments(rows)
}

func (r *Repository) GetByID(ctx context.Context, id, requesterID uuid.UUID) (documentdomain.Document, error) {
	if err := r.ensureDocumentAccess(ctx, id, requesterID); err != nil {
		return documentdomain.Document{}, err
	}

	document, err := r.loadDocumentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return documentdomain.Document{}, documentdomain.ErrDocumentNotFound
		}
		return documentdomain.Document{}, err
	}
	return document, nil
}

func (r *Repository) Create(ctx context.Context, input documentusecase.CreateInput) (documentdomain.Document, error) {
	if input.CreatedBy != nil {
		if err := r.ensureDormitoryAccess(ctx, input.DormitoryID, *input.CreatedBy); err != nil {
			return documentdomain.Document{}, err
		}
	}

	id := uuid.New()
	_, err := r.db.Exec(ctx, `
		INSERT INTO documents (id, dormitory_id, tenant_id, room_id, name, category, file_url, uploaded_date, note, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
	`, id, input.DormitoryID, input.TenantID, input.RoomID, input.Name, input.Category, input.FileURL, input.UploadedDate, input.Note, input.CreatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			switch pgErr.ConstraintName {
			case "documents_tenant_fkey":
				return documentdomain.Document{}, documentdomain.ErrTenantNotFound
			case "documents_room_fkey":
				return documentdomain.Document{}, documentdomain.ErrRoomNotFound
			default:
				return documentdomain.Document{}, documentdomain.ErrDormitoryNotFound
			}
		}
		return documentdomain.Document{}, err
	}

	return r.loadDocumentByID(ctx, id)
}

func (r *Repository) Update(ctx context.Context, id, requesterID uuid.UUID, input documentusecase.UpdateInput) (documentdomain.Document, error) {
	if err := r.ensureDocumentAccess(ctx, id, requesterID); err != nil {
		return documentdomain.Document{}, err
	}

	setClauses := make([]string, 0)
	args := make([]any, 0)
	argIdx := 1

	if input.TenantID != nil {
		setClauses = append(setClauses, fmt.Sprintf("tenant_id = $%d", argIdx))
		args = append(args, *input.TenantID)
		argIdx++
	}
	if input.RoomID != nil {
		setClauses = append(setClauses, fmt.Sprintf("room_id = $%d", argIdx))
		args = append(args, *input.RoomID)
		argIdx++
	}
	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *input.Name)
		argIdx++
	}
	if input.Category != nil {
		setClauses = append(setClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *input.Category)
		argIdx++
	}
	if input.FileURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("file_url = $%d", argIdx))
		args = append(args, *input.FileURL)
		argIdx++
	}
	if input.UploadedDate != nil {
		setClauses = append(setClauses, fmt.Sprintf("uploaded_date = $%d", argIdx))
		args = append(args, *input.UploadedDate)
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
		query := fmt.Sprintf("UPDATE documents SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argIdx)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23503" {
				switch pgErr.ConstraintName {
				case "documents_tenant_fkey":
					return documentdomain.Document{}, documentdomain.ErrTenantNotFound
				case "documents_room_fkey":
					return documentdomain.Document{}, documentdomain.ErrRoomNotFound
				}
			}
			return documentdomain.Document{}, err
		}
	}

	return r.loadDocumentByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id, requesterID uuid.UUID) error {
	if err := r.ensureDocumentAccess(ctx, id, requesterID); err != nil {
		return err
	}

	result, err := r.db.Exec(ctx, `DELETE FROM documents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return documentdomain.ErrDocumentNotFound
	}

	return nil
}

func (r *Repository) loadDocumentByID(ctx context.Context, id uuid.UUID) (documentdomain.Document, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectDocumentColumns+documentFromJoins+` WHERE doc.id = $1`, id)
	return scanDocument(row)
}

func scanDocument(row pgx.Row) (documentdomain.Document, error) {
	var document documentdomain.Document
	if err := row.Scan(
		&document.ID,
		&document.DormitoryID,
		&document.DormitoryName,
		&document.TenantID,
		&document.TenantName,
		&document.RoomID,
		&document.RoomNumber,
		&document.Name,
		&document.Category,
		&document.FileURL,
		&document.UploadedDate,
		&document.Note,
		&document.CreatedBy,
		&document.UpdatedBy,
		&document.CreatedAt,
		&document.UpdatedAt,
	); err != nil {
		return documentdomain.Document{}, err
	}
	return document, nil
}

func scanDocuments(rows pgx.Rows) ([]documentdomain.Document, error) {
	documents := make([]documentdomain.Document, 0)
	for rows.Next() {
		document, err := scanDocument(rows)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return documents, nil
}

// dormitoryScope reports whether the user's role is exempt from per-dormitory
// scoping (sees and manages documents in every dormitory), along with their
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

// ensureDormitoryAccess confirms the dormitory exists and the requester may
// attach documents to it (unrestricted, individually assigned via
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
			return documentdomain.ErrDormitoryNotFound
		}
		return err
	}
	return nil
}

// ensureDocumentAccess confirms the document exists and the requester may
// act on it, based on access to its parent dormitory. Both a missing
// document and a missing grant surface as ErrDocumentNotFound so scoped-out
// callers can't distinguish the two.
func (r *Repository) ensureDocumentAccess(ctx context.Context, id, requesterID uuid.UUID) error {
	full, roleID, err := r.dormitoryScope(ctx, requesterID)
	if err != nil {
		return err
	}

	var exists int
	err = r.db.QueryRow(ctx, `
		SELECT 1 FROM documents doc
		WHERE doc.id = $1
		AND ($2 OR EXISTS (
			SELECT 1 FROM user_dormitories ud WHERE ud.dormitory_id = doc.dormitory_id AND ud.user_id = $3
		) OR EXISTS (
			SELECT 1 FROM role_dormitories rd WHERE rd.dormitory_id = doc.dormitory_id AND rd.role_id = $4
		))
	`, id, full, requesterID, roleID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return documentdomain.ErrDocumentNotFound
		}
		return err
	}
	return nil
}
