package domain

import (
	"time"

	"github.com/google/uuid"
)

// GenerationCandidate is an active contract in a dormitory that could be
// billed for a period, with what the bulk-generation screen needs to warn
// about before creating the invoice.
type GenerationCandidate struct {
	ContractID uuid.UUID  `json:"contract_id"`
	TenantName string     `json:"tenant_name"`
	RoomID     uuid.UUID  `json:"room_id"`
	RoomNumber string     `json:"room_number"`
	RentPrice  float64    `json:"rent_price"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	// AlreadyInvoiced means the period is billed already; generation skips it.
	AlreadyInvoiced bool `json:"already_invoiced"`
	// HasElectricity/HasWater report whether a reading exists in the period.
	// Invoices pull meter charges from the period's readings, so a missing
	// one bills that room without the charge.
	HasElectricity bool `json:"has_electricity"`
	HasWater       bool `json:"has_water"`
	// EndedBeforePeriod flags a contract still active although its end date
	// is before the period starts (the tenant stayed on without renewal).
	EndedBeforePeriod bool `json:"ended_before_period"`
}

// GeneratedInvoice summarises one invoice created by bulk generation.
type GeneratedInvoice struct {
	InvoiceID   uuid.UUID `json:"invoice_id"`
	ContractID  uuid.UUID `json:"contract_id"`
	TenantName  string    `json:"tenant_name"`
	RoomNumber  string    `json:"room_number"`
	TotalAmount float64   `json:"total_amount"`
}

// GenerationFailure is a contract bulk generation could not bill.
type GenerationFailure struct {
	ContractID uuid.UUID `json:"contract_id"`
	TenantName string    `json:"tenant_name"`
	RoomNumber string    `json:"room_number"`
	Error      string    `json:"error"`
}

// GenerationResult reports a bulk generation run. Each invoice is created in
// its own transaction, so a failure on one contract doesn't undo the others.
type GenerationResult struct {
	Created []GeneratedInvoice  `json:"created"`
	Skipped int                 `json:"skipped"`
	Failed  []GenerationFailure `json:"failed"`
}
