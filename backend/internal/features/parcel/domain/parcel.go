package domain

import (
	"context"
	"time"
)

type ParcelStatus string

const (
	ParcelStatusPending  ParcelStatus = "pending"
	ParcelStatusPickedUp ParcelStatus = "picked_up"
)

type Parcel struct {
	ID             string       `json:"id"`
	DormitoryID    string       `json:"dormitory_id"`
	TrackingNumber string       `json:"tracking_number"`
	RecipientName  string       `json:"recipient_name"`
	RoomNumber     string       `json:"room_number"`
	Status         ParcelStatus `json:"status"`
	ReceivedDate   time.Time    `json:"received_date"`
	PickedUpDate   *time.Time   `json:"picked_up_date"`
	Note           string       `json:"note"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type CreateParcelRequest struct {
	TrackingNumber string       `json:"tracking_number"`
	RecipientName  string       `json:"recipient_name"`
	RoomNumber     string       `json:"room_number"`
	Status         ParcelStatus `json:"status"`
	ReceivedDate   time.Time    `json:"received_date"`
	PickedUpDate   *time.Time   `json:"picked_up_date"`
	Note           string       `json:"note"`
}

type UpdateParcelRequest struct {
	TrackingNumber string       `json:"tracking_number"`
	RecipientName  string       `json:"recipient_name"`
	RoomNumber     string       `json:"room_number"`
	Status         ParcelStatus `json:"status"`
	ReceivedDate   time.Time    `json:"received_date"`
	PickedUpDate   *time.Time   `json:"picked_up_date"`
	Note           string       `json:"note"`
}

type ParcelRepository interface {
	FindByID(ctx context.Context, dormitoryID, id string) (*Parcel, error)
	List(ctx context.Context, dormitoryID string, limit, offset int) ([]*Parcel, error)
	Count(ctx context.Context, dormitoryID string) (int, error)
	Create(ctx context.Context, p *Parcel) error
	Update(ctx context.Context, p *Parcel) error
	Delete(ctx context.Context, dormitoryID, id string) error
}
