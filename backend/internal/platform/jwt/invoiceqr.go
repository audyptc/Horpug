package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Invoice QR tokens go into the public image URL of the PromptPay QR sent
// with LINE reminders (LINE fetches images from a public HTTPS URL, so the
// link can't carry a session). They are signed with their own derived key,
// so they never pass as staff or tenant tokens, and only name one invoice.
func invoiceQRKey(secret string) []byte {
	return []byte(secret + "|invoice-qr")
}

type InvoiceQRClaims struct {
	InvoiceID uuid.UUID `json:"inv"`
	jwt.RegisteredClaims
}

func GenerateInvoiceQR(secret string, ttl time.Duration, invoiceID uuid.UUID) (string, error) {
	claims := InvoiceQRClaims{
		InvoiceID: invoiceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(invoiceQRKey(secret))
}

func ParseInvoiceQR(secret, tokenString string) (uuid.UUID, error) {
	claims := &InvoiceQRClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return invoiceQRKey(secret), nil
	})
	if err != nil || !token.Valid || claims.InvoiceID == uuid.Nil {
		return uuid.Nil, ErrInvalidToken
	}
	return claims.InvoiceID, nil
}
