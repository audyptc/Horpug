package domain

import (
	"context"
	"time"
)

type MeterType string

const (
	MeterTypeElectric MeterType = "electric"
	MeterTypeWater    MeterType = "water"
)

type MeterReading struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	MeterType       MeterType `json:"meter_type"`
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type MeterReadingDetail struct {
	MeterReading
	RoomNumber  string  `json:"room_number"`
	UnitUsed    float64 `json:"unit_used"`
	TotalAmount float64 `json:"total_amount"`
}

type CreateMeterReadingRequest struct {
	RoomID          string    `json:"room_id"`
	MeterType       MeterType `json:"meter_type"`
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
}

type UpdateMeterReadingRequest struct {
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
}

type MeterReadingRepository interface {
	FindByID(ctx context.Context, id string) (*MeterReading, error)
	FindDetailByID(ctx context.Context, id string) (*MeterReadingDetail, error)
	List(ctx context.Context, limit, offset int) ([]*MeterReadingDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, m *MeterReading) error
	Update(ctx context.Context, m *MeterReading) error
	Delete(ctx context.Context, id string) error
}
