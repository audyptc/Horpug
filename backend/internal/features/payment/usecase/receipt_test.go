package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	paymentdomain "apihorpug/internal/features/payment/domain"

	"github.com/google/uuid"
)

func TestChangesReceipt(t *testing.T) {
	day := time.Date(2026, time.September, 24, 0, 0, 0, 0, time.UTC)
	before := paymentdomain.Payment{
		PaymentDate: day,
		Items: []paymentdomain.PaymentItem{
			{PaymentMethod: paymentdomain.PaymentMethodCash, Amount: 3000, ReferenceNo: ""},
			{PaymentMethod: paymentdomain.PaymentMethodTransfer, Amount: 2000, ReferenceNo: "TX1"},
		},
	}
	same := func() UpdateInput {
		return UpdateInput{
			PaymentDate: day,
			Note:        "anything",
			Items: []ItemInput{
				{PaymentMethod: paymentdomain.PaymentMethodCash, Amount: 3000},
				{PaymentMethod: paymentdomain.PaymentMethodTransfer, Amount: 2000, ReferenceNo: "TX1"},
			},
		}
	}

	cases := []struct {
		name   string
		mutate func(*UpdateInput)
		want   bool
	}{
		{"note only", func(*UpdateInput) {}, false},
		{"reference corrected", func(in *UpdateInput) { in.Items[1].ReferenceNo = "TX1-fixed" }, false},
		{"float noise in amount", func(in *UpdateInput) { in.Items[0].Amount = 3000.001 }, false},
		{"amount changed", func(in *UpdateInput) { in.Items[0].Amount = 2500 }, true},
		{"method changed", func(in *UpdateInput) { in.Items[0].PaymentMethod = paymentdomain.PaymentMethodOther }, true},
		{"line removed", func(in *UpdateInput) { in.Items = in.Items[:1] }, true},
		{"date changed", func(in *UpdateInput) { in.PaymentDate = day.AddDate(0, 0, 1) }, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := same()
			tc.mutate(&input)
			if got := changesReceipt(before, input); got != tc.want {
				t.Fatalf("changesReceipt() = %v, want %v", got, tc.want)
			}
		})
	}
}

type fakeVoidRepo struct {
	Repository
	payment paymentdomain.Payment
	voided  bool
}

func (f *fakeVoidRepo) GetByID(context.Context, uuid.UUID, uuid.UUID) (paymentdomain.Payment, error) {
	return f.payment, nil
}

func (f *fakeVoidRepo) Void(context.Context, uuid.UUID, uuid.UUID, string) (paymentdomain.InvoiceStatusChange, error) {
	f.voided = true
	return paymentdomain.InvoiceStatusChange{}, nil
}

func TestVoidRejectsAlreadyVoided(t *testing.T) {
	repo := &fakeVoidRepo{payment: paymentdomain.Payment{Status: paymentdomain.PaymentStatusVoided}}
	err := New(repo, nil).Void(context.Background(), uuid.New(), uuid.New(), "dup", "")
	if !errors.Is(err, paymentdomain.ErrPaymentVoided) {
		t.Fatalf("Void() error = %v, want ErrPaymentVoided", err)
	}
	if repo.voided {
		t.Fatal("repository Void called for an already voided payment")
	}
}

func TestUpdateRejectsReceiptChange(t *testing.T) {
	day := time.Date(2026, time.September, 24, 0, 0, 0, 0, time.UTC)
	repo := &fakeVoidRepo{payment: paymentdomain.Payment{
		Status:      paymentdomain.PaymentStatusActive,
		PaymentDate: day,
		Items:       []paymentdomain.PaymentItem{{PaymentMethod: paymentdomain.PaymentMethodCash, Amount: 100}},
	}}
	_, err := New(repo, nil).Update(context.Background(), uuid.New(), uuid.New(), UpdateInput{
		PaymentDate: day,
		Items:       []ItemInput{{PaymentMethod: paymentdomain.PaymentMethodCash, Amount: 200}},
	}, "")
	if !errors.Is(err, paymentdomain.ErrReceiptIssued) {
		t.Fatalf("Update() error = %v, want ErrReceiptIssued", err)
	}
}
