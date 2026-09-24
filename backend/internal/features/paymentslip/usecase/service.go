package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	paymentdomain "apihorpug/internal/features/payment/domain"
	paymentusecase "apihorpug/internal/features/payment/usecase"
	slipdomain "apihorpug/internal/features/paymentslip/domain"

	"github.com/google/uuid"
)

// MaxSlipSize caps a slip image; phone screenshots are well under this.
const MaxSlipSize = 5 << 20

const listLimit = 100

// slipTypes are the image types a slip may be; the stored extension comes
// from here, never from the uploaded file name.
var slipTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type Repository interface {
	PayableInvoiceForTenant(ctx context.Context, tenantID, invoiceID uuid.UUID) (float64, error)
	Create(ctx context.Context, s slipdomain.Slip) (slipdomain.Slip, error)
	ListForTenantInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]slipdomain.Slip, error)
	CancelByTenant(ctx context.Context, tenantID, id uuid.UUID) (slipdomain.Slip, error)
	ListForStaff(ctx context.Context, requesterID uuid.UUID, status slipdomain.Status, limit int) ([]slipdomain.Slip, error)
	GetForStaff(ctx context.Context, id, requesterID uuid.UUID) (slipdomain.Slip, error)
	Claim(ctx context.Context, id, reviewerID uuid.UUID) (bool, error)
	Unclaim(ctx context.Context, id uuid.UUID) error
	SetPayment(ctx context.Context, id, paymentID uuid.UUID) error
	Reject(ctx context.Context, id, reviewerID uuid.UUID, reason string) (bool, error)
	Today(ctx context.Context) (time.Time, error)
}

// Payments records the payment an approved slip stands for, through the
// normal payment flow (receipt number, invoice status, activity log).
type Payments interface {
	Create(ctx context.Context, input paymentusecase.CreateInput, ipAddress string) (paymentdomain.Payment, error)
}

// Notifier tells the tenant on LINE how their slip was handled.
type Notifier interface {
	PushMessage(ctx context.Context, lineUserID, text string) error
}

type Service struct {
	repo     Repository
	files    Files
	payments Payments
	notifier Notifier
}

// Files is the subset of the file store slips use.
type Files interface {
	SaveBytes(folder, ext string, data []byte) (string, error)
	Read(key string) ([]byte, error)
	Delete(key string) error
}

func New(repo Repository, files Files, payments Payments, notifier Notifier) *Service {
	return &Service{repo: repo, files: files, payments: payments, notifier: notifier}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// SubmitInput is a slip as the tenant sends it.
type SubmitInput struct {
	TenantID     uuid.UUID
	InvoiceID    uuid.UUID
	Amount       float64
	TransferDate time.Time
	Note         string
	Image        []byte
}

// validateImage checks the content is one of the allowed image types, by
// sniffing the bytes rather than trusting the file name.
func validateImage(data []byte) (mime, ext string, err error) {
	if len(data) == 0 {
		return "", "", slipdomain.ErrEmptyFile
	}
	if len(data) > MaxSlipSize {
		return "", "", slipdomain.ErrFileTooLarge
	}
	mime = http.DetectContentType(data)
	ext, ok := slipTypes[mime]
	if !ok {
		return "", "", slipdomain.ErrUnsupportedFile
	}
	return mime, ext, nil
}

// Submit stores a tenant's slip for review. The amount may be less than
// what is owed (a part payment) but not more.
func (s *Service) Submit(ctx context.Context, in SubmitInput) (slipdomain.Slip, error) {
	in.Note = strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(in.Note) > 255 {
		in.Note = string([]rune(in.Note)[:255])
	}
	in.Amount = round2(in.Amount)

	mime, ext, err := validateImage(in.Image)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	today, err := s.repo.Today(ctx)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if in.TransferDate.IsZero() || in.TransferDate.After(today) {
		return slipdomain.Slip{}, slipdomain.ErrInvalidDate
	}
	outstanding, err := s.repo.PayableInvoiceForTenant(ctx, in.TenantID, in.InvoiceID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if in.Amount <= 0 || in.Amount > outstanding+0.005 {
		return slipdomain.Slip{}, slipdomain.ErrInvalidAmount
	}

	key, err := s.files.SaveBytes("slips", ext, in.Image)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	slip, err := s.repo.Create(ctx, slipdomain.Slip{
		InvoiceID:    in.InvoiceID,
		TenantID:     in.TenantID,
		Amount:       in.Amount,
		TransferDate: in.TransferDate,
		Note:         in.Note,
		FileKey:      key,
		FileMime:     mime,
		FileSize:     int64(len(in.Image)),
	})
	if err != nil {
		_ = s.files.Delete(key)
		return slipdomain.Slip{}, err
	}
	return slip, nil
}

func (s *Service) ListForTenantInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]slipdomain.Slip, error) {
	return s.repo.ListForTenantInvoice(ctx, tenantID, invoiceID)
}

func (s *Service) CancelByTenant(ctx context.Context, tenantID, id uuid.UUID) (slipdomain.Slip, error) {
	return s.repo.CancelByTenant(ctx, tenantID, id)
}

func (s *Service) ListForStaff(ctx context.Context, requesterID uuid.UUID, status slipdomain.Status) ([]slipdomain.Slip, error) {
	if !status.Valid() {
		status = slipdomain.StatusPending
	}
	return s.repo.ListForStaff(ctx, requesterID, status, listLimit)
}

// Image returns a slip's image for a staff reviewer.
func (s *Service) Image(ctx context.Context, id, requesterID uuid.UUID) (slipdomain.Slip, []byte, error) {
	slip, err := s.repo.GetForStaff(ctx, id, requesterID)
	if err != nil {
		return slipdomain.Slip{}, nil, err
	}
	data, err := s.files.Read(slip.FileKey)
	if err != nil {
		return slipdomain.Slip{}, nil, slipdomain.ErrSlipNotFound
	}
	return slip, data, nil
}

// ApproveInput is what the reviewer confirms; zero values fall back to the
// slip's own amount and transfer date.
type ApproveInput struct {
	Amount      float64
	PaymentDate time.Time
	ReferenceNo string
}

// Approve records the slip as a transfer payment. The slip is claimed first,
// so it can't be approved twice; if recording the payment fails, the claim
// is released and the slip goes back to pending.
func (s *Service) Approve(ctx context.Context, id, reviewerID uuid.UUID, in ApproveInput, ipAddress string) (slipdomain.Slip, error) {
	slip, err := s.repo.GetForStaff(ctx, id, reviewerID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if slip.Status != slipdomain.StatusPending {
		return slipdomain.Slip{}, slipdomain.ErrNotPending
	}
	if in.Amount == 0 {
		in.Amount = slip.Amount
	}
	if in.PaymentDate.IsZero() {
		in.PaymentDate = slip.TransferDate
	}
	in.Amount = round2(in.Amount)
	if in.Amount <= 0 {
		return slipdomain.Slip{}, slipdomain.ErrInvalidAmount
	}

	claimed, err := s.repo.Claim(ctx, id, reviewerID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if !claimed {
		return slipdomain.Slip{}, slipdomain.ErrNotPending
	}

	note := "ชำระผ่านสลิปที่ผู้เช่าส่งทาง LINE"
	if slip.Note != "" {
		note += ": " + slip.Note
	}
	payment, err := s.payments.Create(ctx, paymentusecase.CreateInput{
		InvoiceID:   slip.InvoiceID,
		PaymentDate: in.PaymentDate,
		Note:        note,
		Items: []paymentusecase.ItemInput{{
			PaymentMethod: paymentdomain.PaymentMethodTransfer,
			Amount:        in.Amount,
			ReferenceNo:   strings.TrimSpace(in.ReferenceNo),
		}},
		CreatedBy: &reviewerID,
	}, ipAddress)
	if err != nil {
		if unclaimErr := s.repo.Unclaim(ctx, id); unclaimErr != nil {
			log.Printf("failed to release slip %s after a failed approval: %v", id, unclaimErr)
		}
		return slipdomain.Slip{}, err
	}
	if err := s.repo.SetPayment(ctx, id, payment.ID); err != nil {
		// The payment exists and the slip is approved; only the link is
		// missing, which staff can see from the payment note.
		log.Printf("failed to link slip %s to payment %s: %v", id, payment.ID, err)
	}

	approved, err := s.repo.GetForStaff(ctx, id, reviewerID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	s.notify(ctx, approved, fmt.Sprintf(
		"✅ ได้รับชำระเงินแล้ว\nห้อง %s งวด %02d/%d\nยอด %.2f บาท\nเลขที่ใบเสร็จ %s\nขอบคุณค่ะ 🙏",
		approved.RoomNumber, approved.PeriodMonth, approved.PeriodYear, payment.TotalAmount, payment.ReceiptNo))
	return approved, nil
}

// Reject turns a slip down with a reason the tenant will see.
func (s *Service) Reject(ctx context.Context, id, reviewerID uuid.UUID, reason string) (slipdomain.Slip, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return slipdomain.Slip{}, slipdomain.ErrReasonRequired
	}
	if utf8.RuneCountInString(reason) > 255 {
		reason = string([]rune(reason)[:255])
	}
	if _, err := s.repo.GetForStaff(ctx, id, reviewerID); err != nil {
		return slipdomain.Slip{}, err
	}
	ok, err := s.repo.Reject(ctx, id, reviewerID, reason)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	if !ok {
		return slipdomain.Slip{}, slipdomain.ErrNotPending
	}
	rejected, err := s.repo.GetForStaff(ctx, id, reviewerID)
	if err != nil {
		return slipdomain.Slip{}, err
	}
	s.notify(ctx, rejected, fmt.Sprintf(
		"❌ สลิปการโอนไม่ผ่านการตรวจสอบ\nห้อง %s งวด %02d/%d ยอด %.2f บาท\nเหตุผล: %s\nกรุณาส่งสลิปใหม่ในเมนูผู้เช่า หรือติดต่อเจ้าหน้าที่",
		rejected.RoomNumber, rejected.PeriodMonth, rejected.PeriodYear, rejected.Amount, reason))
	return rejected, nil
}

// notify is best-effort: a tenant without LINE, or a LINE outage, never
// undoes a review.
func (s *Service) notify(ctx context.Context, slip slipdomain.Slip, text string) {
	if s.notifier == nil || slip.TenantLineUserID == "" {
		return
	}
	if err := s.notifier.PushMessage(ctx, slip.TenantLineUserID, text); err != nil {
		log.Printf("failed to notify tenant about slip %s: %v", slip.ID, err)
	}
}
