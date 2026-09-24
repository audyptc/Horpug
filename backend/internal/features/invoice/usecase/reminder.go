package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

// Reminders go out only in daytime Bangkok hours, so a sweep that happens to
// run at night doesn't push to tenants while they sleep; the next daytime
// sweep picks them up.
const (
	reminderStartHour = 9  // inclusive
	reminderEndHour   = 20 // exclusive
)

var bangkok = time.FixedZone("ICT", 7*60*60)

// ReminderRepository finds overdue invoices due a reminder and records attempts.
type ReminderRepository interface {
	ListDueReminders(ctx context.Context) ([]invoicedomain.ReminderCandidate, error)
	RecordReminder(ctx context.Context, invoiceID uuid.UUID, outcome invoicedomain.ReminderOutcome) error
}

// InvoiceJobsRepository is what the hourly invoice jobs need.
type InvoiceJobsRepository interface {
	OverdueMarker
	ReminderRepository
}

// ReminderPusher delivers reminders through the dormitory's LINE OA.
type ReminderPusher interface {
	LinePusher
	PushTextWithImage(ctx context.Context, lineUserID, text, imageURL string) error
	Configured() bool
}

// QRLinker builds the public PromptPay QR image URL for an invoice (see
// NewQRLinker); nil when there is no public HTTPS address to link to.
type QRLinker func(invoiceID uuid.UUID) string

// SendOverdueReminders pushes the weekly LINE reminder for every overdue
// invoice that is due one. A tenant who can't be reached (not a friend of the
// OA, or blocked it) is recorded and retried a week later; a push that fails
// for any other reason (network, LINE outage) is left for the next sweep.
//
// When qrLink is set and the dormitory has a PromptPay account, the PromptPay
// QR image is sent along with the text.
func SendOverdueReminders(ctx context.Context, repo ReminderRepository, pusher ReminderPusher, qrLink QRLinker, now time.Time) {
	if pusher == nil || !pusher.Configured() {
		return
	}
	if hour := now.In(bangkok).Hour(); hour < reminderStartHour || hour >= reminderEndHour {
		return
	}

	candidates, err := repo.ListDueReminders(ctx)
	if err != nil {
		log.Printf("overdue reminder lookup failed: %v", err)
		return
	}

	sent := 0
	for _, c := range candidates {
		if ctx.Err() != nil {
			return
		}

		isFriend, err := pusher.IsFriend(ctx, c.TenantLineUserID)
		if err != nil {
			log.Printf("overdue reminder friendship check failed (invoice=%s): %v", c.InvoiceID, err)
			continue
		}
		if !isFriend {
			if err := repo.RecordReminder(ctx, c.InvoiceID, invoicedomain.ReminderUnreachable); err != nil {
				log.Printf("failed to record overdue reminder (invoice=%s): %v", c.InvoiceID, err)
			}
			continue
		}

		text := buildReminderLineMessage(c, now)
		var imageURL string
		if qrLink != nil && c.PromptPayID != "" && c.TotalAmount > c.PaidAmount {
			imageURL = qrLink(c.InvoiceID)
		}
		if imageURL != "" {
			err = pusher.PushTextWithImage(ctx, c.TenantLineUserID, text, imageURL)
		} else {
			err = pusher.PushMessage(ctx, c.TenantLineUserID, text)
		}
		if err != nil {
			log.Printf("overdue reminder push failed (invoice=%s): %v", c.InvoiceID, err)
			continue
		}
		if err := repo.RecordReminder(ctx, c.InvoiceID, invoicedomain.ReminderSent); err != nil {
			// Delivered but not recorded: it may go out again next sweep,
			// which beats silently losing the record.
			log.Printf("failed to record overdue reminder (invoice=%s): %v", c.InvoiceID, err)
			continue
		}
		sent++
	}

	if sent > 0 {
		log.Printf("overdue reminders: %d sent", sent)
	}
}

// buildReminderLineMessage renders the reminder in the same plain-text style
// as the invoice message (see buildInvoiceLineMessage).
func buildReminderLineMessage(c invoicedomain.ReminderCandidate, now time.Time) string {
	const divider = "－－－－－－－－－－－－"

	outstanding := math.Max(0, math.Round((c.TotalAmount-c.PaidAmount)*100)/100)
	today := now.In(bangkok)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(c.DueDate.Year(), c.DueDate.Month(), c.DueDate.Day(), 0, 0, 0, 0, time.UTC)
	daysLate := int(todayDate.Sub(dueDate).Hours() / 24)

	var b strings.Builder
	b.WriteString("⚠️ แจ้งเตือนค่าเช่าค้างชำระ\n")
	b.WriteString(divider + "\n")
	if c.InvoiceNo != "" {
		fmt.Fprintf(&b, "🧾 เลขที่: %s\n", c.InvoiceNo)
	}
	if c.DormitoryName != "" {
		fmt.Fprintf(&b, "🏢 หอพัก: %s\n", c.DormitoryName)
	}
	if c.RoomNumber != "" {
		fmt.Fprintf(&b, "🚪 ห้อง: %s\n", c.RoomNumber)
	}
	if c.PeriodMonth >= 1 && c.PeriodMonth <= 12 {
		fmt.Fprintf(&b, "📅 งวด: %s %d\n", thaiMonths[c.PeriodMonth], c.PeriodYear+543)
	}
	fmt.Fprintf(&b, "⏰ ครบกำหนดชำระ: %s", c.DueDate.Format("02/01/2006"))
	if daysLate > 0 {
		fmt.Fprintf(&b, " (เลยกำหนด %d วัน)", daysLate)
	}
	b.WriteString("\n")
	b.WriteString(divider + "\n")
	fmt.Fprintf(&b, "💰 ยอดค้างชำระ: %.2f บาท\n", outstanding)
	if c.PromptPayID != "" {
		fmt.Fprintf(&b, "💳 ชำระผ่านพร้อมเพย์: %s\n", formatPromptPayID(c.PromptPayID))
	}
	b.WriteString(divider + "\n")
	b.WriteString("กรุณาชำระโดยเร็ว หากชำระแล้วขออภัยและขอบคุณค่ะ 🙏")

	return b.String()
}

// formatPromptPayID groups a mobile number as 081-234-5678 so it's easy to
// read and type; other IDs are shown as stored.
func formatPromptPayID(id string) string {
	if len(id) == 10 {
		return id[:3] + "-" + id[3:6] + "-" + id[6:]
	}
	return id
}

// RunInvoiceJobs runs the hourly invoice housekeeping until ctx is cancelled:
// it marks overdue invoices, then sends the weekly LINE reminders for them.
// It runs once at start and then every interval.
func RunInvoiceJobs(ctx context.Context, repo InvoiceJobsRepository, pusher ReminderPusher, qrLink QRLinker, interval time.Duration) {
	run := func() {
		jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		sweepOverdue(jobCtx, repo)
		SendOverdueReminders(jobCtx, repo, pusher, qrLink, time.Now())
	}

	run()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}
