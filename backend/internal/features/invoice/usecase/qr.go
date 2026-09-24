package usecase

import (
	"context"
	"math"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	platformjwt "apihorpug/internal/platform/jwt"
	"apihorpug/internal/platform/promptpay"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

// qrImageSize is the PNG's width and height in pixels; large enough for a
// banking app to scan straight off a phone screen.
const qrImageSize = 512

// QRLinkTTL is how long the QR link in a reminder keeps working. Reminders
// repeat weekly, so each one carries a fresh link.
const QRLinkTTL = 30 * 24 * time.Hour

// QRRepository loads what the QR image needs.
type QRRepository interface {
	GetQRInfo(ctx context.Context, id uuid.UUID) (invoicedomain.QRInfo, error)
}

// QRImages serves the PromptPay QR of an invoice as a PNG behind a signed,
// expiring link, for LINE to download when showing a reminder.
type QRImages struct {
	repo   QRRepository
	secret string
}

func NewQRImages(repo QRRepository, secret string) *QRImages {
	return &QRImages{repo: repo, secret: secret}
}

// PNG renders the QR for the amount still owed right now, so a partial
// payment made after the reminder is reflected. Any reason not to show a QR
// (bad or expired token, invoice settled or cancelled, no PromptPay account)
// is ErrInvoiceNotFound, so the public endpoint reveals nothing.
func (q *QRImages) PNG(ctx context.Context, token string) ([]byte, error) {
	id, err := platformjwt.ParseInvoiceQR(q.secret, token)
	if err != nil {
		return nil, invoicedomain.ErrInvoiceNotFound
	}

	info, err := q.repo.GetQRInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	owing := info.Status == invoicedomain.InvoiceStatusUnpaid || info.Status == invoicedomain.InvoiceStatusOverdue
	outstanding := math.Max(0, math.Round((info.TotalAmount-info.PaidAmount)*100)/100)
	if !owing || outstanding <= 0 || info.PromptPayID == "" {
		return nil, invoicedomain.ErrInvoiceNotFound
	}

	payload, err := promptpay.Payload(info.PromptPayID, outstanding)
	if err != nil {
		return nil, invoicedomain.ErrInvoiceNotFound
	}
	return qrcode.Encode(payload, qrcode.Medium, qrImageSize)
}

// NewQRLinker returns a function that builds the public QR image URL for an
// invoice, or nil when publicURL isn't HTTPS: LINE only accepts HTTPS image
// URLs, so reminders then fall back to text only.
func NewQRLinker(publicURL, secret string) func(invoiceID uuid.UUID) string {
	base := strings.TrimRight(strings.TrimSpace(publicURL), "/")
	if !strings.HasPrefix(base, "https://") {
		return nil
	}
	return func(invoiceID uuid.UUID) string {
		token, err := platformjwt.GenerateInvoiceQR(secret, QRLinkTTL, invoiceID)
		if err != nil {
			return ""
		}
		return base + "/api/v1/public/invoice-qr/" + token
	}
}
