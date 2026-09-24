package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

const testSecret = "0123456789abcdef0123456789abcdef"

func TestTenantTokenRoundTrip(t *testing.T) {
	id := uuid.New()
	token, _, err := GenerateTenant(testSecret, time.Hour, id)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := ParseTenant(testSecret, token)
	if err != nil || claims.TenantID != id {
		t.Fatalf("ParseTenant() = %+v, %v; want tenant %s", claims, err, id)
	}
}

func TestTenantAndStaffTokensDoNotCross(t *testing.T) {
	tenantToken, _, err := GenerateTenant(testSecret, time.Hour, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(testSecret, tenantToken); err == nil {
		t.Fatal("a tenant token was accepted as a staff token")
	}

	staffToken, _, err := Generate(testSecret, time.Hour, uuid.New(), uuid.New(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTenant(testSecret, staffToken); err == nil {
		t.Fatal("a staff token was accepted as a tenant token")
	}
}

func TestTenantTokenExpires(t *testing.T) {
	token, _, err := GenerateTenant(testSecret, -time.Minute, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseTenant(testSecret, token); err == nil {
		t.Fatal("an expired tenant token was accepted")
	}
}
