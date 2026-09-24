package usecase

import (
	"fmt"
	"math"
	"strings"
	"time"

	moveoutdomain "apihorpug/internal/features/moveout/domain"

	"github.com/google/uuid"
)

// OutstandingInvoice is an unpaid or overdue invoice of the contract with what
// is still owed on it.
type OutstandingInvoice struct {
	ID          uuid.UUID
	PeriodYear  int
	PeriodMonth int
	Outstanding float64
}

// Reading is a meter reading for the room, taken during the contract up to
// the move-out date, that no invoice or earlier move-out has billed.
type Reading struct {
	ID          uuid.UUID
	Kind        moveoutdomain.ItemType // ItemElectricity or ItemWater
	ReadingDate time.Time
	Amount      float64
}

// Inputs is what the repository gathers for a settlement.
type Inputs struct {
	OutstandingInvoices []OutstandingInvoice
	// MoveOutMonthRentBilled is the rent already invoiced for the move-out
	// month, or nil when that month has no invoice yet.
	MoveOutMonthRentBilled *float64
	UnbilledReadings       []Reading
}

// ManualItem is a deduction staff add (from a preset or typed in).
type ManualItem struct {
	Description string
	Amount      float64
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// normalizeManualItems trims and validates staff-entered deductions.
func normalizeManualItems(items []ManualItem) ([]ManualItem, error) {
	out := make([]ManualItem, 0, len(items))
	for _, item := range items {
		item.Description = strings.TrimSpace(item.Description)
		item.Amount = round2(item.Amount)
		if item.Description == "" || item.Amount <= 0 {
			return nil, moveoutdomain.ErrInvalidItem
		}
		out = append(out, item)
	}
	return out, nil
}

// BuildSettlement works out the deposit reckoning. base carries the contract
// details (deposit, rent, dates); the result adds the lines and totals.
//
// Rent for the move-out month is charged for the days actually stayed,
// counting the move-out day, out of that month's days. If the month was
// already invoiced, the difference is credited back instead of charged.
func BuildSettlement(base moveoutdomain.Settlement, in Inputs, manual []ManualItem) (moveoutdomain.Settlement, error) {
	if base.MoveOutDate.IsZero() {
		return moveoutdomain.Settlement{}, moveoutdomain.ErrRequiredDate
	}
	moveOut := dateOnly(base.MoveOutDate)
	start := dateOnly(base.StartDate)
	if moveOut.Before(start) {
		return moveoutdomain.Settlement{}, moveoutdomain.ErrDateBeforeStart
	}
	manual, err := normalizeManualItems(manual)
	if err != nil {
		return moveoutdomain.Settlement{}, err
	}

	s := base
	s.MoveOutDate = moveOut
	s.Items = make([]moveoutdomain.Item, 0)

	for _, inv := range in.OutstandingInvoices {
		if inv.Outstanding < 0.005 {
			continue
		}
		id := inv.ID
		s.Items = append(s.Items, moveoutdomain.Item{
			ItemType:    moveoutdomain.ItemInvoice,
			Description: fmt.Sprintf("ค้างชำระใบแจ้งหนี้งวด %02d/%d", inv.PeriodMonth, inv.PeriodYear),
			Amount:      round2(inv.Outstanding),
			ReferenceID: &id,
		})
	}

	monthStart := time.Date(moveOut.Year(), moveOut.Month(), 1, 0, 0, 0, 0, time.UTC)
	s.DaysInMonth = monthStart.AddDate(0, 1, -1).Day()
	from := monthStart
	if start.After(from) {
		from = start
	}
	s.DaysStayed = int(moveOut.Sub(from).Hours()/24) + 1
	prorated := round2(base.RentPrice * float64(s.DaysStayed) / float64(s.DaysInMonth))
	period := fmt.Sprintf("%02d/%d", moveOut.Month(), moveOut.Year())

	if in.MoveOutMonthRentBilled == nil {
		if prorated > 0 {
			s.Items = append(s.Items, moveoutdomain.Item{
				ItemType:    moveoutdomain.ItemRent,
				Description: fmt.Sprintf("ค่าเช่างวด %s (%d/%d วัน)", period, s.DaysStayed, s.DaysInMonth),
				Amount:      prorated,
			})
		}
	} else if diff := round2(prorated - *in.MoveOutMonthRentBilled); diff < 0 {
		s.Items = append(s.Items, moveoutdomain.Item{
			ItemType:    moveoutdomain.ItemRentCredit,
			Description: fmt.Sprintf("คืนค่าเช่างวด %s ส่วนที่ไม่ได้พัก (พัก %d/%d วัน)", period, s.DaysStayed, s.DaysInMonth),
			Amount:      diff,
		})
	}

	for _, reading := range in.UnbilledReadings {
		label := "ค่าไฟฟ้า"
		if reading.Kind == moveoutdomain.ItemWater {
			label = "ค่าน้ำประปา"
		}
		id := reading.ID
		s.Items = append(s.Items, moveoutdomain.Item{
			ItemType:    reading.Kind,
			Description: fmt.Sprintf("%s (จดมิเตอร์ %s)", label, reading.ReadingDate.Format("02/01/2006")),
			Amount:      round2(reading.Amount),
			ReferenceID: &id,
		})
	}

	for _, item := range manual {
		s.Items = append(s.Items, moveoutdomain.Item{
			ItemType:    moveoutdomain.ItemOther,
			Description: item.Description,
			Amount:      item.Amount,
		})
	}

	total := 0.0
	for _, item := range s.Items {
		total += item.Amount
	}
	s.TotalDeductions = round2(total)
	net := round2(base.Deposit - s.TotalDeductions)
	s.RefundAmount = math.Max(0, net)
	s.AmountDue = math.Max(0, -net)
	return s, nil
}
