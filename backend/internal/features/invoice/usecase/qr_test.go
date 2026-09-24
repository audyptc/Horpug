package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"
	platformjwt "apihorpug/internal/platform/jwt"

	"github.com/google/uuid"
)

type fakeQRRepo struct{ info invoicedomain.QRInfo }

func (f fakeQRRepo) GetQRInfo(context.Context, uuid.UUID) (invoicedomain.QRInfo, error) {
	return f.info, nil
}

func TestQRImagesPNG(t *testing.T) {
	token, _ := platformjwt.GenerateInvoiceQR("s", time.Hour, uuid.New())
	owing := invoicedomain.QRInfo{Status: invoicedomain.InvoiceStatusOverdue, PromptPayID: "0812345678", TotalAmount: 3000, PaidAmount: 1000}

	png, err := NewQRImages(fakeQRRepo{owing}, "s").PNG(context.Background(), token)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(png, []byte("\x89PNG")) {
		t.Fatal("not a PNG")
	}

	cases := map[string]struct {
		info  invoicedomain.QRInfo
		token string
	}{
		"bad token":    {owing, "nope"},
		"other secret": {owing, func() string { t, _ := platformjwt.GenerateInvoiceQR("x", time.Hour, uuid.New()); return t }()},
		"paid":         {invoicedomain.QRInfo{Status: invoicedomain.InvoiceStatusPaid, PromptPayID: "0812345678", TotalAmount: 3000, PaidAmount: 3000}, token},
		"cancelled":    {invoicedomain.QRInfo{Status: invoicedomain.InvoiceStatusCancelled, PromptPayID: "0812345678", TotalAmount: 3000}, token},
		"no promptpay": {invoicedomain.QRInfo{Status: invoicedomain.InvoiceStatusOverdue, TotalAmount: 3000}, token},
		"fully paid":   {invoicedomain.QRInfo{Status: invoicedomain.InvoiceStatusOverdue, PromptPayID: "0812345678", TotalAmount: 3000, PaidAmount: 3000}, token},
	}
	for name, tc := range cases {
		if _, err := NewQRImages(fakeQRRepo{tc.info}, "s").PNG(context.Background(), tc.token); !errors.Is(err, invoicedomain.ErrInvoiceNotFound) {
			t.Errorf("%s: got %v, want ErrInvoiceNotFound", name, err)
		}
	}
}
