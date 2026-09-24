package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrNotLinked means the LINE account signing in isn't linked to any
	// active tenant, so there is nothing to show.
	ErrNotLinked        = errors.New("this LINE account is not linked to a tenant")
	ErrInvalidLineToken = errors.New("invalid or expired LINE id token")
	ErrRoomNotRented    = errors.New("you can only report repairs for a room you currently rent")
	ErrInvalidRepair    = errors.New("choose a category and describe the problem (up to 255 characters)")
	ErrRepairNotFound   = errors.New("repair request not found")
	ErrRepairNotPending = errors.New("only a pending repair request can be cancelled")
)

// Session is what a tenant gets after signing in from LINE.
type Session struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	Profile     Profile   `json:"profile"`
}

// Room is a room the tenant currently rents (an active contract).
type Room struct {
	ContractID     uuid.UUID `json:"contract_id"`
	RoomID         uuid.UUID `json:"room_id"`
	RoomNumber     string    `json:"room_number"`
	DormitoryID    uuid.UUID `json:"dormitory_id"`
	DormitoryName  string    `json:"dormitory_name"`
	DormitoryPhone string    `json:"dormitory_phone"`
	RentPrice      float64   `json:"rent_price"`
	StartDate      time.Time `json:"start_date"`
}

type Profile struct {
	TenantID  uuid.UUID `json:"tenant_id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Rooms     []Room    `json:"rooms"`
}

// Invoice is a line of the tenant's invoice list.
type Invoice struct {
	ID            uuid.UUID `json:"id"`
	RoomNumber    string    `json:"room_number"`
	DormitoryName string    `json:"dormitory_name"`
	PeriodYear    int       `json:"period_year"`
	PeriodMonth   int       `json:"period_month"`
	DueDate       time.Time `json:"due_date"`
	TotalAmount   float64   `json:"total_amount"`
	Outstanding   float64   `json:"outstanding"`
	Status        string    `json:"status"`
}

type RepairRequest struct {
	ID           uuid.UUID `json:"id"`
	RoomNumber   string    `json:"room_number"`
	Category     string    `json:"category"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	ReportedDate time.Time `json:"reported_date"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Announcement struct {
	ID            uuid.UUID `json:"id"`
	DormitoryName string    `json:"dormitory_name"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	Category      string    `json:"category"`
	IsPinned      bool      `json:"is_pinned"`
	PublishedDate time.Time `json:"published_date"`
}

// RepairCategories are the categories a repair request may have (they match
// the repair_requests check constraint).
var RepairCategories = map[string]bool{
	"electrical": true, "plumbing": true, "furniture": true, "aircon": true, "other": true,
}
