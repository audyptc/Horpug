package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReminderCandidate is an overdue invoice due its weekly LINE reminder, with
// what the message needs.
type ReminderCandidate struct {
	InvoiceID        uuid.UUID
	InvoiceNo        string
	DormitoryID      uuid.UUID
	DormitoryName    string
	PromptPayID      string
	RoomNumber       string
	TenantLineUserID string
	PeriodYear       int
	PeriodMonth      int
	DueDate          time.Time
	TotalAmount      float64
	PaidAmount       float64
	ReminderCount    int
}

// ReminderOutcome is how a reminder attempt ended, as stored in
// invoice_reminders.
type ReminderOutcome string

const (
	ReminderSent ReminderOutcome = "sent"
	// ReminderUnreachable means the tenant hasn't added the OA as a friend
	// or has blocked it; LINE would silently drop the push.
	ReminderUnreachable ReminderOutcome = "unreachable"
)
