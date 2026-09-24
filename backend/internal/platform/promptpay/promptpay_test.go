package promptpay

import (
	"errors"
	"strings"
	"testing"
)

func TestCRC16CheckValue(t *testing.T) {
	// The standard check value for CRC-16/CCITT-FALSE.
	if got := crc16("123456789"); got != 0x29B1 {
		t.Fatalf("crc16 = %04X, want 29B1", got)
	}
}

func TestNormalizeID(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"081-234-5678", "0812345678", false},
		{"1 2345 67890 12 3", "1234567890123", false},
		{"123456789012345", "123456789012345", false},
		{"8123456789", "", true},   // 10 digits not starting with 0
		{"081234567", "", true},    // too short
		{"08123456789a", "", true}, // not a digit
	}
	for _, tc := range cases {
		got, err := NormalizeID(tc.in)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("NormalizeID(%q) = %q, %v; want %q, err=%v", tc.in, got, err, tc.want, tc.wantErr)
		}
	}
}

func TestPayload(t *testing.T) {
	cases := []struct {
		name   string
		id     string
		amount float64
		body   string
	}{
		{
			name:   "mobile, static",
			id:     "0812345678",
			amount: 0,
			body:   "000201010211" + "29370016A000000677010111" + "01130066812345678" + "5802TH" + "5303764" + "6304",
		},
		{
			name:   "mobile with amount",
			id:     "0812345678",
			amount: 4200.5,
			body:   "000201010212" + "29370016A000000677010111" + "01130066812345678" + "5802TH" + "5303764" + "54074200.50" + "6304",
		},
		{
			name:   "national id",
			id:     "1234567890123",
			amount: 100,
			body:   "000201010212" + "29370016A000000677010111" + "02131234567890123" + "5802TH" + "5303764" + "5406100.00" + "6304",
		},
		{
			name:   "e-wallet",
			id:     "123456789012345",
			amount: 0,
			body:   "000201010211" + "29390016A000000677010111" + "0315123456789012345" + "5802TH" + "5303764" + "6304",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Payload(tc.id, tc.amount)
			if err != nil {
				t.Fatalf("Payload() error = %v", err)
			}
			if !strings.HasPrefix(got, tc.body) || len(got) != len(tc.body)+4 {
				t.Fatalf("Payload() = %s, want %s + CRC", got, tc.body)
			}
			if crc := got[len(got)-4:]; crc != strings.ToUpper(crc) {
				t.Fatalf("CRC %s must be uppercase hex", crc)
			}
			if want := crc16(tc.body); got[len(got)-4:] != sprintfCRC(want) {
				t.Fatalf("CRC = %s, want %s", got[len(got)-4:], sprintfCRC(want))
			}
		})
	}
}

func TestPayloadRejectsMissingID(t *testing.T) {
	if _, err := Payload("", 100); !errors.Is(err, ErrInvalidID) {
		t.Fatalf("Payload(\"\") error = %v, want ErrInvalidID", err)
	}
}

func sprintfCRC(v uint16) string {
	const hex = "0123456789ABCDEF"
	return string([]byte{hex[v>>12&0xF], hex[v>>8&0xF], hex[v>>4&0xF], hex[v&0xF]})
}

// Reference payloads produced by the promptpay-qr npm library (v0.5.0), which
// Thai banking apps are known to accept.
func TestPayloadMatchesReferenceLibrary(t *testing.T) {
	cases := []struct {
		id     string
		amount float64
		want   string
	}{
		{"0812345678", 0, "00020101021129370016A000000677010111011300668123456785802TH530376463045D82"},
		{"0812345678", 4200.5, "00020101021229370016A000000677010111011300668123456785802TH530376454074200.506304BF95"},
		{"1234567890123", 100, "00020101021229370016A000000677010111021312345678901235802TH53037645406100.006304BB6C"},
		{"123456789012345", 0, "00020101021129390016A00000067701011103151234567890123455802TH5303764630473AF"},
	}
	for _, tc := range cases {
		got, err := Payload(tc.id, tc.amount)
		if err != nil {
			t.Fatalf("Payload(%s, %v) error = %v", tc.id, tc.amount, err)
		}
		if got != tc.want {
			t.Errorf("Payload(%s, %v) = %s, want %s", tc.id, tc.amount, got, tc.want)
		}
	}
}
