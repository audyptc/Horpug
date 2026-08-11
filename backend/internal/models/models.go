package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string       `json:"name" gorm:"size:120;uniqueIndex;not null"`
	Description string       `json:"description" gorm:"size:255"`
	IsActive    bool         `json:"is_active" gorm:"not null;default:true"`
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string    `json:"name" gorm:"size:120;uniqueIndex;not null"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Username  string     `json:"username" gorm:"size:80;uniqueIndex;not null"`
	Email     string     `json:"email" gorm:"size:180;uniqueIndex;not null"`
	Password  string     `json:"-" gorm:"size:255;not null"`
	RoleID    *uuid.UUID `json:"role_id" gorm:"type:uuid"`
	Role      *Role      `json:"role,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	IsActive  bool       `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (m *Role) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *Permission) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *User) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
