package usecase

import (
	"context"
	"time"

	dashboarddomain "apihorpug/internal/features/dashboard/domain"

	"github.com/google/uuid"
)

// ExpiryWindowDays is how far ahead a contract's end date counts as "expiring
// soon".
const ExpiryWindowDays = 30

// businessZone decides which calendar day and month the dashboard is talking
// about. The database session runs in UTC, which would flip the month seven
// hours late for the dormitories this serves; Thailand has no daylight saving,
// so a fixed offset is exact.
var businessZone = time.FixedZone("ICT", 7*60*60)

// Window is the set of dates the summary's time-based figures are measured
// against. They are calendar dates, carried as midnight UTC so they reach the
// database as plain DATE values whatever the server's own zone is.
type Window struct {
	Today          time.Time
	MonthStart     time.Time
	NextMonthStart time.Time
	ExpiryLimit    time.Time
	Year           int
	Month          int
}

// NewWindow derives the dates from an instant, reading it in the business zone.
func NewWindow(now time.Time) Window {
	local := now.In(businessZone)
	year, month, day := local.Date()

	monthStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return Window{
		Today:          today,
		MonthStart:     monthStart,
		NextMonthStart: monthStart.AddDate(0, 1, 0),
		ExpiryLimit:    today.AddDate(0, 0, ExpiryWindowDays),
		Year:           year,
		Month:          int(month),
	}
}

type Repository interface {
	GetSummary(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID, window Window) (dashboarddomain.Summary, error)
}

type Service struct {
	repo Repository
	now  func() time.Time
}

func New(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now}
}

// GetSummary counts over the dormitories the requester may access, or over one
// of them when dormitoryID is set.
func (s *Service) GetSummary(ctx context.Context, requesterID uuid.UUID, dormitoryID *uuid.UUID) (dashboarddomain.Summary, error) {
	return s.repo.GetSummary(ctx, requesterID, dormitoryID, NewWindow(s.now()))
}
