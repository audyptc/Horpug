package domain

import (
	"time"

	roledomain "apihorpug/internal/features/role/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey"`
	Username  string           `json:"username" gorm:"size:80;uniqueIndex;not null"`
	Email     string           `json:"email" gorm:"size:180;uniqueIndex;not null"`
	Password  string           `json:"-" gorm:"size:255;not null"`
	RoleID    uuid.UUID        `json:"role_id" gorm:"type:uuid;not null"`
	Role      *roledomain.Role `json:"role,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT"`
	IsActive  bool             `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (m *User) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
