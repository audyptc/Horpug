package domain

import (
	"context"
	"time"
)

type ElectricBillingType string

const (
	ElectricBillingTypeMeter ElectricBillingType = "meter"
	ElectricBillingTypeFlat  ElectricBillingType = "flat"
)

type ElectricMeter struct {
	ID              string              `json:"id"`
	RoomID          string              `json:"room_id"`
	BillingType     ElectricBillingType `json:"billing_type"`
	BillingMonth    *time.Time          `json:"billing_month"`
	ReadingDate     time.Time           `json:"reading_date"`
	PreviousReading float64             `json:"previous_reading"`
	CurrentReading  *float64            `json:"current_reading"`
	UnitPrice       *float64            `json:"unit_price"`
	FlatAmount      *float64            `json:"flat_amount"`
	Note            string              `json:"note"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	CreatedBy       string              `json:"created_by"`
	UpdatedBy       string              `json:"updated_by"`
	UpdatedByName   string              `json:"updated_by_name"`
}

type ElectricMeterDetail struct {
	ElectricMeter
	RoomNumber  string   `json:"room_number"`
	UnitUsed    *float64 `json:"unit_used"`
	TotalAmount float64  `json:"total_amount"`
}

type CreateElectricMeterRequest struct {
	RoomID          string              `json:"room_id"`
	BillingType     ElectricBillingType `json:"billing_type"`
	BillingMonth    *time.Time          `json:"billing_month"`
	ReadingDate     time.Time           `json:"reading_date"`
	PreviousReading float64             `json:"previous_reading"`
	CurrentReading  *float64            `json:"current_reading"`
	UnitPrice       *float64            `json:"unit_price"`
	FlatAmount      *float64            `json:"flat_amount"`
	Note            string              `json:"note"`
}

type UpdateElectricMeterRequest struct {
	BillingType     ElectricBillingType `json:"billing_type"`
	BillingMonth    *time.Time          `json:"billing_month"`
	ReadingDate     time.Time           `json:"reading_date"`
	PreviousReading float64             `json:"previous_reading"`
	CurrentReading  *float64            `json:"current_reading"`
	UnitPrice       *float64            `json:"unit_price"`
	FlatAmount      *float64            `json:"flat_amount"`
	Note            string              `json:"note"`
}

type ElectricMeterRepository interface {
	FindByID(ctx context.Context, id string) (*ElectricMeter, error)
	FindDetailByID(ctx context.Context, id string) (*ElectricMeterDetail, error)
	FindLatestByRoomID(ctx context.Context, roomID string) (*ElectricMeterDetail, error)
	List(ctx context.Context, limit, offset int) ([]*ElectricMeterDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, m *ElectricMeter) error
	Update(ctx context.Context, m *ElectricMeter) error
	Delete(ctx context.Context, id string) error
}
