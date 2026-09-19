package sqlutil

import "testing"

func TestContainsPattern(t *testing.T) {
	cases := map[string]string{
		"":         "%%",
		"abc":      "%abc%",
		"50%":      `%50\%%`,
		"a_b":      `%a\_b%`,
		`a\b`:      `%a\\b%`,
		`\`:        `%\\%`,
		`\%`:       `%\\\%%`,
		"100% _ok": `%100\% \_ok%`,
		"ห้อง":     "%ห้อง%",
	}
	for input, want := range cases {
		if got := ContainsPattern(input); got != want {
			t.Errorf("ContainsPattern(%q) = %q, want %q", input, got, want)
		}
	}
}
