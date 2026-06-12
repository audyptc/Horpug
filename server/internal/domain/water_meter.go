package domain

import (
	"context"
	"time"
)

type WaterBillingType string

const (
	WaterBillingTypeMeter WaterBillingType = "meter"
	WaterBillingTypeFlat  WaterBillingType = "flat"
)

type WaterMeter struct {
	ID              string           `json:"id"`
	RoomID          string           `json:"room_id"`
	BillingType     WaterBillingType `json:"billing_type"`
	ReadingDate     time.Time        `json:"reading_date"`
	PreviousReading *float64         `json:"previous_reading"`
	CurrentReading  *float64         `json:"current_reading"`
	UnitPrice       *float64         `json:"unit_price"`
	FlatAmount      *float64         `json:"flat_amount"`
	Note            string           `json:"note"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	CreatedBy       string           `json:"created_by"`
	UpdatedBy       string           `json:"updated_by"`
	UpdatedByName   string           `json:"updated_by_name"`
}

type WaterMeterDetail struct {
	WaterMeter
	RoomNumber  string   `json:"room_number"`
	UnitUsed    *float64 `json:"unit_used"`
	TotalAmount float64  `json:"total_amount"`
}

type CreateWaterMeterRequest struct {
	RoomID          string           `json:"room_id"`
	BillingType     WaterBillingType `json:"billing_type"`
	ReadingDate     time.Time        `json:"reading_date"`
	PreviousReading *float64         `json:"previous_reading"`
	CurrentReading  *float64         `json:"current_reading"`
	UnitPrice       *float64         `json:"unit_price"`
	FlatAmount      *float64         `json:"flat_amount"`
	Note            string           `json:"note"`
}

type UpdateWaterMeterRequest struct {
	BillingType     WaterBillingType `json:"billing_type"`
	ReadingDate     time.Time        `json:"reading_date"`
	PreviousReading *float64         `json:"previous_reading"`
	CurrentReading  *float64         `json:"current_reading"`
	UnitPrice       *float64         `json:"unit_price"`
	FlatAmount      *float64         `json:"flat_amount"`
	Note            string           `json:"note"`
}

type WaterMeterRepository interface {
	FindByID(ctx context.Context, id string) (*WaterMeter, error)
	FindDetailByID(ctx context.Context, id string) (*WaterMeterDetail, error)
	List(ctx context.Context, limit, offset int) ([]*WaterMeterDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, m *WaterMeter) error
	Update(ctx context.Context, m *WaterMeter) error
	Delete(ctx context.Context, id string) error
}
