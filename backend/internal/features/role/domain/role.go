package domain

import (
	"time"

	menudomain "apihorpug/internal/features/menu/domain"
	permissiondomain "apihorpug/internal/features/permission/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role struct {
	ID              uuid.UUID            `json:"id" gorm:"type:uuid;primaryKey"`
	Name            string               `json:"name" gorm:"size:120;uniqueIndex;not null"`
	Description     string               `json:"description" gorm:"size:255"`
	IsActive        bool                 `json:"is_active" gorm:"not null;default:true"`
	MenuPermissions []RoleMenuPermission `json:"menu_permissions,omitempty" gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

type RoleMenuPermission struct {
	RoleID       uuid.UUID                   `json:"role_id" gorm:"type:uuid;primaryKey"`
	MenuID       uuid.UUID                   `json:"menu_id" gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID                   `json:"permission_id" gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time                   `json:"created_at"`
	Menu         menudomain.Menu             `json:"menu" gorm:"foreignKey:MenuID;references:ID;constraint:OnDelete:CASCADE"`
	Permission   permissiondomain.Permission `json:"permission" gorm:"foreignKey:PermissionID;references:ID;constraint:OnDelete:CASCADE"`
}

func (m *Role) BeforeCreate(_ *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
