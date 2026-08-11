package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name        string    `json:"name" gorm:"size:120;uniqueIndex;not null"`
	Description string    `json:"description" gorm:"size:255"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Action is the closed set of central permission names. Actions are generic
// (not tied to a resource) and get scoped to a resource by binding them to a
// menu via role_menu_permissions. Add new actions here to extend the set.
type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

// Actions lists every valid Action, in seed/display order.
var Actions = []Action{ActionCreate, ActionRead, ActionUpdate, ActionDelete}

func (a Action) Valid() bool {
	for _, action := range Actions {
		if action == a {
			return true
		}
	}
	return false
}
