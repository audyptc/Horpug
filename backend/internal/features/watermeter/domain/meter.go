package domain

import (
	"time"

	"github.com/google/uuid"
)

type BillingMethod string

const (
	// BillingMethodMetered derives total_amount from (current_unit -
	// previous_unit) * price_per_unit. Use for rooms with a real per-room
	// meter, including cases where price_per_unit is a blended rate computed
	// externally (e.g. from a tiered/progressive water rate).
	BillingMethodMetered BillingMethod = "metered"
	// BillingMethodFlat uses flat_amount directly as total_amount. Use when
	// there's no per-room reading to bill against: flat monthly charges, or
	// amounts allocated from a shared/central meter and computed externally.
	BillingMethodFlat BillingMethod = "flat"
)

func (m BillingMethod) Valid() bool {
	switch m {
	case BillingMethodMetered, BillingMethodFlat:
		return true
	}
	return false
}

type Meter struct {
	ID            uuid.UUID     `json:"id"`
	RoomID        uuid.UUID     `json:"room_id"`
	RoomNumber    string        `json:"room_number,omitempty"`
	DormitoryID   uuid.UUID     `json:"dormitory_id,omitempty"`
	DormitoryName string        `json:"dormitory_name,omitempty"`
	BillingMethod BillingMethod `json:"billing_method"`
	ReadingDate   time.Time     `json:"reading_date"`
	PreviousUnit  float64       `json:"previous_unit"`
	CurrentUnit   float64       `json:"current_unit"`
	UnitUsed      float64       `json:"unit_used"`
	PricePerUnit  float64       `json:"price_per_unit"`
	FlatAmount    *float64      `json:"flat_amount,omitempty"`
	TotalAmount   float64       `json:"total_amount"`
	IsBilled      bool          `json:"is_billed"`
	Note          string        `json:"note"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID    `json:"updated_by,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}
