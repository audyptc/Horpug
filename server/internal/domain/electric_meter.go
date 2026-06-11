package domain

import (
	"context"
	"time"
)

type ElectricMeter struct {
	ID              string    `json:"id"`
	RoomID          string    `json:"room_id"`
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ElectricMeterDetail struct {
	ElectricMeter
	RoomNumber  string  `json:"room_number"`
	UnitUsed    float64 `json:"unit_used"`
	TotalAmount float64 `json:"total_amount"`
}

type CreateElectricMeterRequest struct {
	RoomID          string    `json:"room_id"`
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
}

type UpdateElectricMeterRequest struct {
	ReadingDate     time.Time `json:"reading_date"`
	PreviousReading float64   `json:"previous_reading"`
	CurrentReading  float64   `json:"current_reading"`
	UnitPrice       float64   `json:"unit_price"`
	Note            string    `json:"note"`
}

type ElectricMeterRepository interface {
	FindByID(ctx context.Context, id string) (*ElectricMeter, error)
	FindDetailByID(ctx context.Context, id string) (*ElectricMeterDetail, error)
	List(ctx context.Context, limit, offset int) ([]*ElectricMeterDetail, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, m *ElectricMeter) error
	Update(ctx context.Context, m *ElectricMeter) error
	Delete(ctx context.Context, id string) error
}
