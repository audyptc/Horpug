package domain

import (
	"time"

	"github.com/google/uuid"
)

type RoomStatus string

const (
	RoomStatusAvailable   RoomStatus = "available"
	RoomStatusOccupied    RoomStatus = "occupied"
	RoomStatusMaintenance RoomStatus = "maintenance"
)

func (s RoomStatus) Valid() bool {
	switch s {
	case RoomStatusAvailable, RoomStatusOccupied, RoomStatusMaintenance:
		return true
	}
	return false
}

type Room struct {
	ID            uuid.UUID  `json:"id"`
	DormitoryID   uuid.UUID  `json:"dormitory_id"`
	DormitoryName string     `json:"dormitory_name,omitempty"`
	RoomTypeID    uuid.UUID  `json:"room_type_id"`
	RoomTypeName  string     `json:"room_type_name,omitempty"`
	RoomTypePrice float64    `json:"room_type_price"`
	RoomNumber    string     `json:"room_number"`
	Floor         int        `json:"floor"`
	Status        RoomStatus `json:"status"`
	IsActive      bool       `json:"is_active"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	UpdatedBy     *uuid.UUID `json:"updated_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
