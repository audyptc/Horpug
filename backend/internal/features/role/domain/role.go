package domain

import (
	"time"

	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID          uuid.UUID                     `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string                        `json:"name" gorm:"size:120;uniqueIndex;not null"`
	Description string                        `json:"description" gorm:"size:255"`
	IsActive    bool                          `json:"is_active" gorm:"not null;default:true"`
	Permissions []permissiondomain.Permission `json:"permissions" gorm:"many2many:role_permissions;constraint:OnDelete:CASCADE"`
	CreatedAt   time.Time                     `json:"created_at"`
	UpdatedAt   time.Time                     `json:"updated_at"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"role_id" gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `json:"permission_id" gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time `json:"created_at"`
}

func (m *Role) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
