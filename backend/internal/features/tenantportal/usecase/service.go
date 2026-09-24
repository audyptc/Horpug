package usecase

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	paymentdomain "apihorpug/internal/features/payment/domain"
	portaldomain "apihorpug/internal/features/tenantportal/domain"
	platformjwt "apihorpug/internal/platform/jwt"

	"github.com/google/uuid"
)

const (
	// A LIFF page is short-lived; when the token runs out the page signs in
	// again silently from LINE.
	sessionTTL = 2 * time.Hour
	listLimit  = 36
)

type Repository interface {
	FindTenantByLineUserID(ctx context.Context, lineUserID string) (uuid.UUID, error)
	Profile(ctx context.Context, tenantID uuid.UUID) (portaldomain.Profile, error)
	Invoices(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.Invoice, error)
	RepairRequests(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.RepairRequest, error)
	CreateRepairRequest(ctx context.Context, tenantID, roomID uuid.UUID, category, description string) (portaldomain.RepairRequest, error)
	CancelRepairRequest(ctx context.Context, tenantID, id uuid.UUID) (portaldomain.RepairRequest, error)
	Announcements(ctx context.Context, tenantID uuid.UUID, limit int) ([]portaldomain.Announcement, error)
}

// LineVerifier checks a LIFF id token and returns the LINE userId it was
// issued for.
type LineVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (string, error)
}

// InvoiceDocuments renders a tenant's own invoice (see the invoice feature).
type InvoiceDocuments interface {
	GetDocumentForTenant(ctx context.Context, id, tenantID uuid.UUID) (invoicedomain.Document, error)
}

// Receipts renders a receipt for one of the tenant's own payments (see the
// payment feature).
type Receipts interface {
	GetReceiptForTenant(ctx context.Context, id, tenantID uuid.UUID) (paymentdomain.Receipt, error)
}

type Service struct {
	repo      Repository
	line      LineVerifier
	documents InvoiceDocuments
	receipts  Receipts
	secret    string
}

func New(repo Repository, line LineVerifier, documents InvoiceDocuments, receipts Receipts, secret string) *Service {
	return &Service{repo: repo, line: line, documents: documents, receipts: receipts, secret: secret}
}

// StartSession signs a tenant in from the LIFF page: LINE vouches for who
// they are (the id token), and the account must be linked to an active
// tenant. No password is involved.
func (s *Service) StartSession(ctx context.Context, idToken string) (portaldomain.Session, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return portaldomain.Session{}, portaldomain.ErrInvalidLineToken
	}
	lineUserID, err := s.line.VerifyIDToken(ctx, idToken)
	if err != nil {
		return portaldomain.Session{}, portaldomain.ErrInvalidLineToken
	}
	tenantID, err := s.repo.FindTenantByLineUserID(ctx, lineUserID)
	if err != nil {
		return portaldomain.Session{}, err
	}
	profile, err := s.repo.Profile(ctx, tenantID)
	if err != nil {
		return portaldomain.Session{}, err
	}
	token, expiresAt, err := platformjwt.GenerateTenant(s.secret, sessionTTL, tenantID)
	if err != nil {
		return portaldomain.Session{}, err
	}
	return portaldomain.Session{AccessToken: token, ExpiresAt: expiresAt, Profile: profile}, nil
}

func (s *Service) Profile(ctx context.Context, tenantID uuid.UUID) (portaldomain.Profile, error) {
	return s.repo.Profile(ctx, tenantID)
}

func (s *Service) Invoices(ctx context.Context, tenantID uuid.UUID) ([]portaldomain.Invoice, error) {
	return s.repo.Invoices(ctx, tenantID, listLimit)
}

func (s *Service) InvoiceDocument(ctx context.Context, tenantID, invoiceID uuid.UUID) (invoicedomain.Document, error) {
	return s.documents.GetDocumentForTenant(ctx, invoiceID, tenantID)
}

func (s *Service) Receipt(ctx context.Context, tenantID, paymentID uuid.UUID) (paymentdomain.Receipt, error) {
	return s.receipts.GetReceiptForTenant(ctx, paymentID, tenantID)
}

func (s *Service) RepairRequests(ctx context.Context, tenantID uuid.UUID) ([]portaldomain.RepairRequest, error) {
	return s.repo.RepairRequests(ctx, tenantID, listLimit)
}

func (s *Service) CreateRepairRequest(ctx context.Context, tenantID, roomID uuid.UUID, category, description string) (portaldomain.RepairRequest, error) {
	description = strings.TrimSpace(description)
	if !portaldomain.RepairCategories[category] || description == "" || utf8.RuneCountInString(description) > 255 || roomID == uuid.Nil {
		return portaldomain.RepairRequest{}, portaldomain.ErrInvalidRepair
	}
	return s.repo.CreateRepairRequest(ctx, tenantID, roomID, category, description)
}

func (s *Service) CancelRepairRequest(ctx context.Context, tenantID, id uuid.UUID) (portaldomain.RepairRequest, error) {
	return s.repo.CancelRepairRequest(ctx, tenantID, id)
}

func (s *Service) Announcements(ctx context.Context, tenantID uuid.UUID) ([]portaldomain.Announcement, error) {
	return s.repo.Announcements(ctx, tenantID, listLimit)
}
