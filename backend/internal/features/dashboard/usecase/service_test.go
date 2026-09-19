package usecase

import (
	"testing"
	"time"
)

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestNewWindow(t *testing.T) {
	cases := []struct {
		name string
		now  time.Time
		want Window
	}{
		{
			name: "midday",
			now:  time.Date(2026, time.April, 15, 5, 0, 0, 0, time.UTC),
			want: Window{
				Today:          date(2026, time.April, 15),
				MonthStart:     date(2026, time.April, 1),
				NextMonthStart: date(2026, time.May, 1),
				ExpiryLimit:    date(2026, time.May, 15),
				Year:           2026,
				Month:          4,
			},
		},
		{
			// 18:00 UTC on 31 March is already 01:00 on 1 April in Thailand, so the
			// dashboard has to have rolled over to April.
			name: "month rolls over seven hours before UTC does",
			now:  time.Date(2026, time.March, 31, 18, 0, 0, 0, time.UTC),
			want: Window{
				Today:          date(2026, time.April, 1),
				MonthStart:     date(2026, time.April, 1),
				NextMonthStart: date(2026, time.May, 1),
				ExpiryLimit:    date(2026, time.May, 1),
				Year:           2026,
				Month:          4,
			},
		},
		{
			name: "december to january",
			now:  time.Date(2026, time.December, 20, 0, 0, 0, 0, time.UTC),
			want: Window{
				Today:          date(2026, time.December, 20),
				MonthStart:     date(2026, time.December, 1),
				NextMonthStart: date(2027, time.January, 1),
				ExpiryLimit:    date(2027, time.January, 19),
				Year:           2026,
				Month:          12,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := NewWindow(tc.now); got != tc.want {
				t.Errorf("NewWindow(%s)\n got %+v\nwant %+v", tc.now, got, tc.want)
			}
		})
	}
}
