package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Tenant tokens are for the tenant self-service pages opened from LINE. They
// are signed with a key derived from the app secret but different from the
// staff key, so a tenant token can never pass as a staff token (Parse
// rejects it) and a staff token never passes ParseTenant.
func tenantKey(secret string) []byte {
	return []byte(secret + "|tenant-portal")
}

type TenantClaims struct {
	TenantID uuid.UUID `json:"tenant_id"`
	jwt.RegisteredClaims
}

func GenerateTenant(secret string, ttl time.Duration, tenantID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := TenantClaims{
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tenantKey(secret))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func ParseTenant(secret, tokenString string) (*TenantClaims, error) {
	claims := &TenantClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tenantKey(secret), nil
	})
	if err != nil || !token.Valid || claims.TenantID == uuid.Nil {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
