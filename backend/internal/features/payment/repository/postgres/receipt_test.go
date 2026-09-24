package postgres

import "testing"

func TestFormatReceiptNo(t *testing.T) {
	cases := map[[2]int]string{
		{2026, 1}:     "RC2026-0001",
		{2026, 42}:    "RC2026-0042",
		{2027, 9999}:  "RC2027-9999",
		{2027, 12345}: "RC2027-12345",
	}
	for in, want := range cases {
		if got := FormatReceiptNo(in[0], in[1]); got != want {
			t.Errorf("FormatReceiptNo(%d, %d) = %s, want %s", in[0], in[1], got, want)
		}
	}
}
