package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

// ListFilters narrows and orders an invoice listing. Search casts a wide OR
// across tenant name, room number and dormitory name; Columns narrows
// individual columns by substring on top of that, keyed by FilterColumns.
type ListFilters struct {
	ContractID  *uuid.UUID
	RoomID      *uuid.UUID
	DormitoryID *uuid.UUID
	TenantID    *uuid.UUID
	Status      *invoicedomain.InvoiceStatus
	PeriodYear  *int
	PeriodMonth *int
	Search      string
	Columns     map[string]string
	SortKey     string
	SortDesc    bool
}

// FilterColumns whitelists the columns a caller may match a substring
// against. Same rule as SortColumns: the expression is interpolated into SQL
// rather than bound, so nothing outside this map may reach the query.
var FilterColumns = map[string]string{
	"tenant_name":    "(t.first_name || ' ' || t.last_name)",
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
}

// SortColumns maps the sort keys the API accepts onto the expressions they
// order by. A column name can't be passed to Postgres as a bind parameter, so
// it is interpolated into the query — every value that reaches ORDER BY must
// come from this map and never straight from the request. period combines
// year and month into one comparable value so a single ASC/DESC direction
// orders both correctly.
var SortColumns = map[string]string{
	"tenant_name":    "(t.first_name || ' ' || t.last_name)",
	"room_number":    "rm.room_number",
	"dormitory_name": "d.name",
	"invoice_no":     "i.invoice_no",
	"period":         "(i.period_year * 12 + i.period_month)",
	"due_date":       "i.due_date",
	"total_amount":   "i.total_amount",
	"status":         "i.status",
	"created_at":     "i.created_at",
}

const DefaultSortKey = "period"

type CreateInput struct {
	ContractID  uuid.UUID
	PeriodYear  int
	PeriodMonth int
	IssueDate   time.Time
	DueDate     time.Time
	Note        string
	CreatedBy   *uuid.UUID
}

type UpdateInput struct {
	Status    *invoicedomain.InvoiceStatus
	DueDate   *time.Time
	Note      *string
	UpdatedBy *uuid.UUID
}

type AddItemInput struct {
	Description string
	Amount      float64
	UpdatedBy   *uuid.UUID
}

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]invoicedomain.Invoice, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Invoice, error)
	Create(ctx context.Context, input CreateInput) (invoicedomain.Invoice, error)
	Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput) (invoicedomain.Invoice, error)
	Delete(ctx context.Context, id, requesterID uuid.UUID) error
	AddItem(ctx context.Context, invoiceID, requesterID uuid.UUID, input AddItemInput) (invoicedomain.Invoice, error)
	RemoveItem(ctx context.Context, invoiceID, itemID, requesterID uuid.UUID) (invoicedomain.Invoice, error)
	GenerationRepository
	DocumentRepository
}

// LinePusher sends a text message to a tenant's linked LINE account through
// the dormitory's LINE Official Account.
type LinePusher interface {
	PushMessage(ctx context.Context, lineUserID, text string) error
	IsFriend(ctx context.Context, lineUserID string) (bool, error)
}

// ActivityLogger records invoice create/update/delete, line-item and LINE-send
// events for the audit trail. Failures to record are logged but never block
// the invoice flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo        Repository
	linePusher  LinePusher
	activityLog ActivityLogger
}

func New(repo Repository, linePusher LinePusher, activityLog ActivityLogger) *Service {
	return &Service{repo: repo, linePusher: linePusher, activityLog: activityLog}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the invoice flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID, dormitoryID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	var dormitoryRef *uuid.UUID
	if dormitoryID != uuid.Nil {
		dormitoryRef = &dormitoryID
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "invoice",
		EntityID:    &entityID,
		DormitoryID: dormitoryRef,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func invoiceActivityDescription(invoice invoicedomain.Invoice) string {
	return fmt.Sprintf("invoice: room %s (%04d-%02d)", invoice.RoomNumber, invoice.PeriodYear, invoice.PeriodMonth)
}

func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]invoicedomain.Invoice, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	invoices, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return invoices, total, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (invoicedomain.Invoice, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (invoicedomain.Invoice, error) {
	input.Note = strings.TrimSpace(input.Note)

	if input.ContractID == uuid.Nil || input.PeriodYear <= 0 || input.IssueDate.IsZero() || input.DueDate.IsZero() {
		return invoicedomain.Invoice{}, invoicedomain.ErrRequiredInvoiceData
	}
	if input.PeriodMonth < 1 || input.PeriodMonth > 12 {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoicePeriod
	}
	if input.DueDate.Before(input.IssueDate) {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceDates
	}

	invoice, err := s.repo.Create(ctx, input)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", invoice.ID, invoice.DormitoryID, "Created "+invoiceActivityDescription(invoice), ipAddress)
	return invoice, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (invoicedomain.Invoice, error) {
	if input.Status != nil && !input.Status.Valid() {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceStatus
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	invoice, err := s.repo.Update(ctx, id, requesterID, input)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}

	description := "Updated " + invoiceActivityDescription(invoice)
	if input.Status != nil {
		description += fmt.Sprintf(" — status: %s", *input.Status)
	}
	s.recordActivity(ctx, &requesterID, "UPDATE", invoice.ID, invoice.DormitoryID, description, ipAddress)
	return invoice, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	invoice, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id, requesterID); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, invoice.DormitoryID, "Deleted "+invoiceActivityDescription(invoice), ipAddress)
	return nil
}

func (s *Service) AddItem(ctx context.Context, invoiceID, requesterID uuid.UUID, input AddItemInput, ipAddress string) (invoicedomain.Invoice, error) {
	input.Description = strings.TrimSpace(input.Description)

	if input.Description == "" {
		return invoicedomain.Invoice{}, invoicedomain.ErrRequiredInvoiceItemData
	}
	if input.Amount <= 0 {
		return invoicedomain.Invoice{}, invoicedomain.ErrInvalidInvoiceItemAmount
	}

	invoice, err := s.repo.AddItem(ctx, invoiceID, requesterID, input)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", invoice.ID, invoice.DormitoryID,
		fmt.Sprintf("Added item %q (%.2f) to %s", input.Description, input.Amount, invoiceActivityDescription(invoice)), ipAddress)
	return invoice, nil
}

func (s *Service) RemoveItem(ctx context.Context, invoiceID, itemID, requesterID uuid.UUID, ipAddress string) (invoicedomain.Invoice, error) {
	invoice, err := s.repo.RemoveItem(ctx, invoiceID, itemID, requesterID)
	if err != nil {
		return invoicedomain.Invoice{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", invoice.ID, invoice.DormitoryID, "Removed item from "+invoiceActivityDescription(invoice), ipAddress)
	return invoice, nil
}

// SendLine pushes a text summary of the invoice to the tenant's linked LINE
// account through the dormitory's LINE Official Account.
func (s *Service) SendLine(ctx context.Context, invoiceID, requesterID uuid.UUID, ipAddress string) error {
	invoice, err := s.repo.GetByID(ctx, invoiceID, requesterID)
	if err != nil {
		return err
	}
	if invoice.TenantLineUserID == "" {
		return invoicedomain.ErrTenantLineNotLinked
	}

	isFriend, err := s.linePusher.IsFriend(ctx, invoice.TenantLineUserID)
	if err != nil {
		return err
	}
	if !isFriend {
		return invoicedomain.ErrTenantLineUnreachable
	}

	if err := s.linePusher.PushMessage(ctx, invoice.TenantLineUserID, buildInvoiceLineMessage(invoice)); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "SEND_LINE", invoice.ID, invoice.DormitoryID, "Sent LINE message for "+invoiceActivityDescription(invoice), ipAddress)
	return nil
}

// LineMessagePreview returns the same text SendLine would push, for tenants
// who only have a personal LINE ID on file (not linked via the OA/LIFF
// flow) — staff can copy this into a manual LINE chat instead.
func (s *Service) LineMessagePreview(ctx context.Context, invoiceID, requesterID uuid.UUID) (string, error) {
	invoice, err := s.repo.GetByID(ctx, invoiceID, requesterID)
	if err != nil {
		return "", err
	}

	return buildInvoiceLineMessage(invoice), nil
}

var thaiMonths = [...]string{
	"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}

// buildInvoiceLineMessage renders the invoice as a plain-text LINE message,
// using line breaks, separators and emoji to approximate a receipt-style
// layout since LINE's push text messages carry no rich formatting.
func buildInvoiceLineMessage(invoice invoicedomain.Invoice) string {
	const divider = "－－－－－－－－－－－－"

	var b strings.Builder

	b.WriteString("🧾 ใบแจ้งหนี้ค่าเช่าหอพัก\n")
	b.WriteString(divider + "\n")
	if invoice.InvoiceNo != "" {
		fmt.Fprintf(&b, "🔖 เลขที่: %s\n", invoice.InvoiceNo)
	}
	if invoice.DormitoryName != "" {
		fmt.Fprintf(&b, "🏢 หอพัก: %s\n", invoice.DormitoryName)
	}
	if invoice.RoomNumber != "" {
		fmt.Fprintf(&b, "🚪 ห้อง: %s\n", invoice.RoomNumber)
	}
	fmt.Fprintf(&b, "📅 งวด: %s %d\n", thaiMonths[invoice.PeriodMonth], invoice.PeriodYear+543)
	b.WriteString(divider + "\n")

	for _, item := range invoice.Items {
		fmt.Fprintf(&b, "• %s: %.2f บาท\n", item.Description, item.Amount)
	}

	b.WriteString(divider + "\n")
	fmt.Fprintf(&b, "💰 ยอดรวมทั้งสิ้น: %.2f บาท\n", invoice.TotalAmount)
	fmt.Fprintf(&b, "⏰ ครบกำหนดชำระ: %s\n", invoice.DueDate.Format("02/01/2006"))
	b.WriteString(divider + "\n")
	b.WriteString("กรุณาชำระภายในกำหนดวันที่แจ้ง ขอบคุณค่ะ 🙏")

	return b.String()
}
