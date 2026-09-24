package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInvoiceQRToken(t *testing.T) {
	id := uuid.New()
	token, err := GenerateInvoiceQR("secret", time.Hour, id)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := ParseInvoiceQR("secret", token); err != nil || got != id {
		t.Fatalf("round trip: got %v, %v", got, err)
	}
	if _, err := ParseInvoiceQR("other", token); err == nil {
		t.Fatal("token signed with another secret must be rejected")
	}

	expired, _ := GenerateInvoiceQR("secret", -time.Minute, id)
	if _, err := ParseInvoiceQR("secret", expired); err == nil {
		t.Fatal("expired token must be rejected")
	}

	// Staff and tenant tokens must not pass as QR tokens.
	staff, _, _ := Generate("secret", time.Hour, id, uuid.New(), "admin")
	if _, err := ParseInvoiceQR("secret", staff); err == nil {
		t.Fatal("staff token must be rejected")
	}
	tenant, _, _ := GenerateTenant("secret", time.Hour, id)
	if _, err := ParseInvoiceQR("secret", tenant); err == nil {
		t.Fatal("tenant token must be rejected")
	}
}
