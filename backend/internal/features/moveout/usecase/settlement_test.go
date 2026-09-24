package usecase

import (
	"errors"
	"testing"
	"time"

	moveoutdomain "apihorpug/internal/features/moveout/domain"

	"github.com/google/uuid"
)

func day(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func baseSettlement(moveOut time.Time) moveoutdomain.Settlement {
	return moveoutdomain.Settlement{
		StartDate:   day(2026, time.January, 1),
		MoveOutDate: moveOut,
		RentPrice:   3000,
		Deposit:     6000,
	}
}

func itemsOfType(s moveoutdomain.Settlement, t moveoutdomain.ItemType) []moveoutdomain.Item {
	var out []moveoutdomain.Item
	for _, item := range s.Items {
		if item.ItemType == t {
			out = append(out, item)
		}
	}
	return out
}

func TestProratedRentWhenMonthNotInvoiced(t *testing.T) {
	// 10 of September's 30 days.
	s, err := BuildSettlement(baseSettlement(day(2026, time.September, 10)), Inputs{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	rent := itemsOfType(s, moveoutdomain.ItemRent)
	if len(rent) != 1 || rent[0].Amount != 1000 {
		t.Fatalf("rent items = %+v, want one of 1000", rent)
	}
	if s.DaysStayed != 10 || s.DaysInMonth != 30 {
		t.Errorf("days = %d/%d, want 10/30", s.DaysStayed, s.DaysInMonth)
	}
	if s.RefundAmount != 5000 || s.AmountDue != 0 {
		t.Errorf("refund/due = %v/%v, want 5000/0", s.RefundAmount, s.AmountDue)
	}
}

func TestRentCreditWhenMonthAlreadyInvoiced(t *testing.T) {
	billed := 3000.0
	s, err := BuildSettlement(baseSettlement(day(2026, time.September, 10)), Inputs{MoveOutMonthRentBilled: &billed}, nil)
	if err != nil {
		t.Fatal(err)
	}
	credit := itemsOfType(s, moveoutdomain.ItemRentCredit)
	if len(credit) != 1 || credit[0].Amount != -2000 {
		t.Fatalf("credit items = %+v, want one of -2000", credit)
	}
	if len(itemsOfType(s, moveoutdomain.ItemRent)) != 0 {
		t.Error("charged rent for a month that was already invoiced")
	}
	if s.RefundAmount != 8000 {
		t.Errorf("refund = %v, want 8000 (deposit + credit)", s.RefundAmount)
	}
}

func TestNoCreditForFullMonth(t *testing.T) {
	billed := 3000.0
	s, err := BuildSettlement(baseSettlement(day(2026, time.September, 30)), Inputs{MoveOutMonthRentBilled: &billed}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != 0 {
		t.Fatalf("items = %+v, want none when the whole month was stayed and billed", s.Items)
	}
}

func TestProrationStartsAtContractStartInSameMonth(t *testing.T) {
	base := baseSettlement(day(2026, time.September, 20))
	base.StartDate = day(2026, time.September, 11)
	s, err := BuildSettlement(base, Inputs{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if s.DaysStayed != 10 {
		t.Errorf("days stayed = %d, want 10 (11th to 20th)", s.DaysStayed)
	}
}

func TestDeductionsExceedingDepositGiveAmountDue(t *testing.T) {
	billed := 3000.0
	s, err := BuildSettlement(baseSettlement(day(2026, time.September, 30)), Inputs{
		MoveOutMonthRentBilled: &billed,
		OutstandingInvoices: []OutstandingInvoice{
			{ID: uuid.New(), PeriodYear: 2026, PeriodMonth: 8, Outstanding: 3500},
			{ID: uuid.New(), PeriodYear: 2026, PeriodMonth: 9, Outstanding: 0.001}, // noise, skipped
		},
		UnbilledReadings: []Reading{
			{ID: uuid.New(), Kind: moveoutdomain.ItemElectricity, ReadingDate: day(2026, time.September, 30), Amount: 420},
			{ID: uuid.New(), Kind: moveoutdomain.ItemWater, ReadingDate: day(2026, time.September, 30), Amount: 80},
		},
	}, []ManualItem{{Description: " ค่าซ่อมประตู ", Amount: 2500}})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != 4 {
		t.Fatalf("items = %+v, want invoice, electricity, water, repair", s.Items)
	}
	if s.TotalDeductions != 6500 || s.RefundAmount != 0 || s.AmountDue != 500 {
		t.Errorf("total/refund/due = %v/%v/%v, want 6500/0/500", s.TotalDeductions, s.RefundAmount, s.AmountDue)
	}
	if other := itemsOfType(s, moveoutdomain.ItemOther); other[0].Description != "ค่าซ่อมประตู" {
		t.Errorf("manual description not trimmed: %q", other[0].Description)
	}
}

func TestBuildSettlementValidation(t *testing.T) {
	cases := []struct {
		name   string
		base   moveoutdomain.Settlement
		manual []ManualItem
		want   error
	}{
		{"no date", baseSettlement(time.Time{}), nil, moveoutdomain.ErrRequiredDate},
		{"before start", baseSettlement(day(2025, time.December, 31)), nil, moveoutdomain.ErrDateBeforeStart},
		{"blank deduction", baseSettlement(day(2026, time.September, 1)), []ManualItem{{Description: " ", Amount: 10}}, moveoutdomain.ErrInvalidItem},
		{"zero deduction", baseSettlement(day(2026, time.September, 1)), []ManualItem{{Description: "x", Amount: 0}}, moveoutdomain.ErrInvalidItem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := BuildSettlement(tc.base, Inputs{}, tc.manual); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}
