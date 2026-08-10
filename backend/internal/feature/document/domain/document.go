package domain

import (
	"context"
	"time"
)

type DocumentCategory string

const (
	DocumentCategoryContract          DocumentCategory = "contract"
	DocumentCategoryIDCard            DocumentCategory = "id_card"
	DocumentCategoryHouseRegistration DocumentCategory = "house_registration"
	DocumentCategoryReceipt           DocumentCategory = "receipt"
	DocumentCategoryOther             DocumentCategory = "other"
)

type Document struct {
	ID         string           `json:"id"`
	Title      string           `json:"title"`
	Category   DocumentCategory `json:"category"`
	TenantID   *string          `json:"tenant_id"`
	FileURL    string           `json:"file_url"`
	IssueDate  *time.Time       `json:"issue_date"`
	ExpiryDate *time.Time       `json:"expiry_date"`
	Note       string           `json:"note"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type DocumentDetail struct {
	Document
	TenantFirstName string `json:"tenant_first_name"`
	TenantLastName  string `json:"tenant_last_name"`
}

type CreateDocumentRequest struct {
	Title      string           `json:"title"`
	Category   DocumentCategory `json:"category"`
	TenantID   *string          `json:"tenant_id"`
	FileURL    string           `json:"file_url"`
	IssueDate  *time.Time       `json:"issue_date"`
	ExpiryDate *time.Time       `json:"expiry_date"`
	Note       string           `json:"note"`
}

type UpdateDocumentRequest struct {
	Title      string           `json:"title"`
	Category   DocumentCategory `json:"category"`
	TenantID   *string          `json:"tenant_id"`
	FileURL    string           `json:"file_url"`
	IssueDate  *time.Time       `json:"issue_date"`
	ExpiryDate *time.Time       `json:"expiry_date"`
	Note       string           `json:"note"`
}

type DocumentRepository interface {
	FindByID(ctx context.Context, id string) (*DocumentDetail, error)
	List(ctx context.Context, limit, offset int) ([]*DocumentDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, d *Document) error
	Update(ctx context.Context, d *Document) error
	Delete(ctx context.Context, id string) error
}
