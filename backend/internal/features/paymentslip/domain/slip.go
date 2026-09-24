package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusApproved  Status = "approved"
	StatusRejected  Status = "rejected"
	StatusCancelled Status = "cancelled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected, StatusCancelled:
		return true
	}
	return false
}

var (
	ErrSlipNotFound      = errors.New("slip not found")
	ErrInvoiceNotPayable = errors.New("this invoice is not waiting for payment")
	ErrInvalidAmount     = errors.New("amount must be greater than zero and not more than what is owed")
	ErrInvalidDate       = errors.New("transfer_date is required and cannot be in the future")
	ErrNotPending        = errors.New("this slip has already been reviewed or cancelled")
	ErrReasonRequired    = errors.New("a reason is required to reject a slip")
	ErrEmptyFile         = errors.New("the slip image is empty")
	ErrFileTooLarge      = errors.New("the slip image is too large")
	ErrUnsupportedFile   = errors.New("the slip must be a JPG, PNG or WEBP image")
)

// Slip is a transfer slip a tenant sent for one of their invoices.
type Slip struct {
	ID            uuid.UUID  `json:"id"`
	InvoiceID     uuid.UUID  `json:"invoice_id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	TenantName    string     `json:"tenant_name"`
	RoomNumber    string     `json:"room_number"`
	DormitoryID   uuid.UUID  `json:"dormitory_id"`
	DormitoryName string     `json:"dormitory_name"`
	PeriodYear    int        `json:"period_year"`
	PeriodMonth   int        `json:"period_month"`
	InvoiceTotal  float64    `json:"invoice_total"`
	Outstanding   float64    `json:"outstanding"`
	Amount        float64    `json:"amount"`
	TransferDate  time.Time  `json:"transfer_date"`
	Note          string     `json:"note"`
	FileMime      string     `json:"file_mime"`
	FileSize      int64      `json:"file_size"`
	Status        Status     `json:"status"`
	RejectReason  string     `json:"reject_reason,omitempty"`
	PaymentID     *uuid.UUID `json:"payment_id,omitempty"`
	ReceiptNo     string     `json:"receipt_no,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`

	// Not sent to clients: where the image is stored, and where to notify
	// the tenant.
	FileKey          string `json:"-"`
	TenantLineUserID string `json:"-"`
}
