package usecase

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"testing"
	"time"

	paymentdomain "apihorpug/internal/features/payment/domain"
	paymentusecase "apihorpug/internal/features/payment/usecase"
	slipdomain "apihorpug/internal/features/paymentslip/domain"

	"github.com/google/uuid"
)

var today = time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC)

type fakeRepo struct {
	Repository
	outstanding float64
	slip        slipdomain.Slip
	claimed     int
	unclaimed   int
	created     *slipdomain.Slip
}

func (f *fakeRepo) Today(context.Context) (time.Time, error) { return today, nil }
func (f *fakeRepo) PayableInvoiceForTenant(context.Context, uuid.UUID, uuid.UUID) (float64, error) {
	return f.outstanding, nil
}
func (f *fakeRepo) Create(_ context.Context, s slipdomain.Slip) (slipdomain.Slip, error) {
	f.created = &s
	return s, nil
}
func (f *fakeRepo) GetForStaff(context.Context, uuid.UUID, uuid.UUID) (slipdomain.Slip, error) {
	return f.slip, nil
}
func (f *fakeRepo) Claim(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	f.claimed++
	if f.slip.Status != slipdomain.StatusPending {
		return false, nil
	}
	f.slip.Status = slipdomain.StatusApproved
	return true, nil
}
func (f *fakeRepo) Unclaim(context.Context, uuid.UUID) error {
	f.unclaimed++
	f.slip.Status = slipdomain.StatusPending
	return nil
}
func (f *fakeRepo) SetPayment(context.Context, uuid.UUID, uuid.UUID) error { return nil }

type fakeFiles struct{ saved, deleted int }

func (f *fakeFiles) SaveBytes(string, string, []byte) (string, error) {
	f.saved++
	return "slips/x.png", nil
}
func (f *fakeFiles) Read(string) ([]byte, error) { return nil, nil }
func (f *fakeFiles) Delete(string) error         { f.deleted++; return nil }

type fakePayments struct {
	err   error
	calls int
	last  paymentusecase.CreateInput
}

func (f *fakePayments) Create(_ context.Context, in paymentusecase.CreateInput, _ string) (paymentdomain.Payment, error) {
	f.calls++
	f.last = in
	if f.err != nil {
		return paymentdomain.Payment{}, f.err
	}
	return paymentdomain.Payment{ID: uuid.New(), ReceiptNo: "RC2026-0001", TotalAmount: in.Items[0].Amount}, nil
}

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestSubmitValidation(t *testing.T) {
	good := func() SubmitInput {
		return SubmitInput{TenantID: uuid.New(), InvoiceID: uuid.New(), Amount: 1000, TransferDate: today, Image: pngBytes(t)}
	}
	cases := []struct {
		name   string
		mutate func(*SubmitInput)
		want   error
	}{
		{"not an image", func(in *SubmitInput) { in.Image = []byte("%PDF-1.4 not an image") }, slipdomain.ErrUnsupportedFile},
		{"empty", func(in *SubmitInput) { in.Image = nil }, slipdomain.ErrEmptyFile},
		{"future date", func(in *SubmitInput) { in.TransferDate = today.AddDate(0, 0, 1) }, slipdomain.ErrInvalidDate},
		{"more than owed", func(in *SubmitInput) { in.Amount = 3000.01 }, slipdomain.ErrInvalidAmount},
		{"zero", func(in *SubmitInput) { in.Amount = 0 }, slipdomain.ErrInvalidAmount},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files := &fakeFiles{}
			svc := New(&fakeRepo{outstanding: 3000}, files, &fakePayments{}, nil)
			in := good()
			tc.mutate(&in)
			if _, err := svc.Submit(context.Background(), in); !errors.Is(err, tc.want) {
				t.Fatalf("Submit() error = %v, want %v", err, tc.want)
			}
			if files.saved != 0 {
				t.Error("stored a file for a rejected slip")
			}
		})
	}

	repo := &fakeRepo{outstanding: 3000}
	svc := New(repo, &fakeFiles{}, &fakePayments{}, nil)
	if _, err := svc.Submit(context.Background(), good()); err != nil {
		t.Fatalf("valid slip rejected: %v", err)
	}
	if repo.created == nil || repo.created.FileMime != "image/png" {
		t.Fatalf("created = %+v, want a png slip", repo.created)
	}
}

func TestApproveRecordsPaymentOnce(t *testing.T) {
	repo := &fakeRepo{slip: slipdomain.Slip{ID: uuid.New(), InvoiceID: uuid.New(), Amount: 1500, TransferDate: today, Status: slipdomain.StatusPending}}
	payments := &fakePayments{}
	svc := New(repo, &fakeFiles{}, payments, nil)

	if _, err := svc.Approve(context.Background(), repo.slip.ID, uuid.New(), ApproveInput{ReferenceNo: "TX9"}, ""); err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if payments.calls != 1 || payments.last.Items[0].Amount != 1500 ||
		payments.last.Items[0].PaymentMethod != paymentdomain.PaymentMethodTransfer || !payments.last.PaymentDate.Equal(today) {
		t.Fatalf("payment = %+v, want one transfer of 1500 dated the transfer day", payments.last)
	}

	if _, err := svc.Approve(context.Background(), repo.slip.ID, uuid.New(), ApproveInput{}, ""); !errors.Is(err, slipdomain.ErrNotPending) {
		t.Fatalf("second Approve() error = %v, want ErrNotPending", err)
	}
	if payments.calls != 1 {
		t.Fatalf("payments recorded = %d, want 1", payments.calls)
	}
}

func TestApproveReleasesClaimWhenPaymentFails(t *testing.T) {
	repo := &fakeRepo{slip: slipdomain.Slip{ID: uuid.New(), Amount: 5000, TransferDate: today, Status: slipdomain.StatusPending}}
	svc := New(repo, &fakeFiles{}, &fakePayments{err: paymentdomain.ErrPaymentExceedsInvoice}, nil)

	_, err := svc.Approve(context.Background(), repo.slip.ID, uuid.New(), ApproveInput{}, "")
	if !errors.Is(err, paymentdomain.ErrPaymentExceedsInvoice) {
		t.Fatalf("Approve() error = %v, want ErrPaymentExceedsInvoice", err)
	}
	if repo.unclaimed != 1 || repo.slip.Status != slipdomain.StatusPending {
		t.Fatalf("slip status = %s after failed approval, want pending again", repo.slip.Status)
	}
}
