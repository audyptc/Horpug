package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

type fakeReminderRepo struct {
	candidates []invoicedomain.ReminderCandidate
	recorded   map[uuid.UUID]invoicedomain.ReminderOutcome
}

func (f *fakeReminderRepo) ListDueReminders(context.Context) ([]invoicedomain.ReminderCandidate, error) {
	return f.candidates, nil
}

func (f *fakeReminderRepo) RecordReminder(_ context.Context, id uuid.UUID, outcome invoicedomain.ReminderOutcome) error {
	f.recorded[id] = outcome
	return nil
}

type fakePusher struct {
	configured bool
	friends    map[string]bool
	failPush   map[string]bool
	pushed     map[string]string
}

func (f *fakePusher) Configured() bool { return f.configured }

func (f *fakePusher) IsFriend(_ context.Context, userID string) (bool, error) {
	return f.friends[userID], nil
}

func (f *fakePusher) PushMessage(_ context.Context, userID, text string) error {
	if f.failPush[userID] {
		return errors.New("line down")
	}
	f.pushed[userID] = text
	return nil
}

// 10:00 in Bangkok.
var daytime = time.Date(2026, time.September, 24, 3, 0, 0, 0, time.UTC)

func TestSendOverdueReminders(t *testing.T) {
	friend, blocked, flaky := uuid.New(), uuid.New(), uuid.New()
	repo := &fakeReminderRepo{
		candidates: []invoicedomain.ReminderCandidate{
			{InvoiceID: friend, TenantLineUserID: "U-friend", PeriodYear: 2026, PeriodMonth: 8, TotalAmount: 3000},
			{InvoiceID: blocked, TenantLineUserID: "U-blocked", PeriodYear: 2026, PeriodMonth: 8},
			{InvoiceID: flaky, TenantLineUserID: "U-flaky", PeriodYear: 2026, PeriodMonth: 8},
		},
		recorded: map[uuid.UUID]invoicedomain.ReminderOutcome{},
	}
	pusher := &fakePusher{
		configured: true,
		friends:    map[string]bool{"U-friend": true, "U-flaky": true},
		failPush:   map[string]bool{"U-flaky": true},
		pushed:     map[string]string{},
	}

	SendOverdueReminders(context.Background(), repo, pusher, daytime)

	if repo.recorded[friend] != invoicedomain.ReminderSent {
		t.Errorf("friend: recorded %q, want sent", repo.recorded[friend])
	}
	if repo.recorded[blocked] != invoicedomain.ReminderUnreachable {
		t.Errorf("blocked: recorded %q, want unreachable (so it waits a week)", repo.recorded[blocked])
	}
	if _, ok := repo.recorded[flaky]; ok {
		t.Errorf("flaky: recorded %q, want nothing so the next sweep retries", repo.recorded[flaky])
	}
	if _, ok := pusher.pushed["U-blocked"]; ok {
		t.Error("pushed to a tenant who isn't a friend of the OA")
	}
}

func TestSendOverdueRemindersSkips(t *testing.T) {
	cases := []struct {
		name       string
		configured bool
		now        time.Time
	}{
		{"line not configured", false, daytime},
		{"before 09:00 Bangkok", true, time.Date(2026, time.September, 24, 1, 59, 0, 0, time.UTC)}, // 08:59
		{"from 20:00 Bangkok", true, time.Date(2026, time.September, 24, 13, 0, 0, 0, time.UTC)},   // 20:00
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeReminderRepo{
				candidates: []invoicedomain.ReminderCandidate{{InvoiceID: uuid.New(), TenantLineUserID: "U1"}},
				recorded:   map[uuid.UUID]invoicedomain.ReminderOutcome{},
			}
			pusher := &fakePusher{configured: tc.configured, friends: map[string]bool{"U1": true}, pushed: map[string]string{}}
			SendOverdueReminders(context.Background(), repo, pusher, tc.now)
			if len(pusher.pushed) != 0 || len(repo.recorded) != 0 {
				t.Fatalf("sent %d, recorded %d; want nothing", len(pusher.pushed), len(repo.recorded))
			}
		})
	}
}

func TestBuildReminderLineMessage(t *testing.T) {
	msg := buildReminderLineMessage(invoicedomain.ReminderCandidate{
		DormitoryName: "หอพัก1",
		RoomNumber:    "102",
		PeriodYear:    2026,
		PeriodMonth:   8,
		DueDate:       time.Date(2026, time.September, 5, 0, 0, 0, 0, time.UTC),
		TotalAmount:   4200,
		PaidAmount:    1200,
		PromptPayID:   "0812345678",
	}, daytime)

	for _, want := range []string{
		"ห้อง: 102",
		"สิงหาคม 2569",
		"05/09/2026 (เลยกำหนด 19 วัน)",
		"ยอดค้างชำระ: 3000.00 บาท",
		"พร้อมเพย์: 081-234-5678",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q:\n%s", want, msg)
		}
	}
}
