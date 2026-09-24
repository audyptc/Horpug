// Package promptpay builds Thai PromptPay QR payloads (EMVCo merchant-presented
// QR, as read by every Thai banking app) for a dormitory's PromptPay account.
package promptpay

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidID means the value isn't a PromptPay target: a 10-digit mobile
// number starting with 0, a 13-digit national/tax ID, or a 15-digit e-wallet ID.
var ErrInvalidID = errors.New("promptpay id must be a 10-digit mobile number, 13-digit national/tax id or 15-digit e-wallet id")

const (
	aidPromptPay  = "A000000677010111"
	tagMobile     = "01"
	tagNationalID = "02"
	tagEWallet    = "03"
)

// NormalizeID strips the separators people type (spaces, dashes, dots) and
// checks what's left is a PromptPay target. An empty input is valid and
// returns "" (no PromptPay account set).
func NormalizeID(raw string) (string, error) {
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '.':
		default:
			return "", ErrInvalidID
		}
	}

	id := b.String()
	switch len(id) {
	case 0:
		return "", nil
	case 10:
		if id[0] != '0' {
			return "", ErrInvalidID
		}
	case 13, 15:
	default:
		return "", ErrInvalidID
	}
	return id, nil
}

// Payload returns the QR payload paying amount (in baht) to id, which must
// already be normalized. A zero amount gives a static QR where the payer
// types the amount; otherwise the amount is fixed in the QR.
func Payload(id string, amount float64) (string, error) {
	id, err := NormalizeID(id)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", ErrInvalidID
	}
	if amount < 0 {
		return "", errors.New("amount must not be negative")
	}

	var target string
	switch len(id) {
	case 10:
		// Mobile numbers go in international form, zero-padded to 13 digits:
		// 0812345678 -> 0066812345678.
		target = field(tagMobile, fmt.Sprintf("%013s", "66"+id[1:]))
	case 13:
		target = field(tagNationalID, id)
	case 15:
		target = field(tagEWallet, id)
	}

	pointOfInitiation := "11" // static, reusable
	if amount > 0 {
		pointOfInitiation = "12" // dynamic, amount fixed
	}

	var b strings.Builder
	b.WriteString(field("00", "01"))
	b.WriteString(field("01", pointOfInitiation))
	b.WriteString(field("29", field("00", aidPromptPay)+target))
	// Same field order as the widely used promptpay-qr library, which is what
	// Thai banking apps are tested against.
	b.WriteString(field("58", "TH"))
	b.WriteString(field("53", "764")) // THB
	if amount > 0 {
		b.WriteString(field("54", fmt.Sprintf("%.2f", amount)))
	}
	b.WriteString("6304")
	b.WriteString(fmt.Sprintf("%04X", crc16(b.String())))
	return b.String(), nil
}

// field encodes one EMVCo TLV: 2-char tag, 2-digit length, value.
func field(tag, value string) string {
	return fmt.Sprintf("%s%02d%s", tag, len(value), value)
}

// crc16 is CRC-16/CCITT-FALSE (poly 0x1021, init 0xFFFF), which EMVCo QR
// uses over the whole payload including the "6304" CRC tag and length.
func crc16(data string) uint16 {
	crc := uint16(0xFFFF)
	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for bit := 0; bit < 8; bit++ {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
