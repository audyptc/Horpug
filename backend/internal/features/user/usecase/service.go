package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	userdomain "apihorpug/internal/features/user/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type DeletionCheck struct {
	CanDelete   bool  `json:"can_delete"`
	RecordCount int64 `json:"record_count"`
	IsProtected bool  `json:"is_protected"`
}

type UserPermissionItem struct {
	MenuID         uuid.UUID `json:"menu_id"`
	MenuName       string    `json:"menu_name"`
	MenuPath       string    `json:"menu_path"`
	PermissionID   uuid.UUID `json:"permission_id"`
	PermissionName string    `json:"permission_name"`
}

type CreateInput struct {
	Username  string
	Email     string
	Password  string
	RoleID    uuid.UUID
	IsActive  bool
	CreatedBy *uuid.UUID
}

type UpdateInput struct {
	Username  *string
	Email     *string
	Password  *string
	RoleID    *uuid.UUID
	IsActive  *bool
	UpdatedBy *uuid.UUID
}

// ListFilters narrows and orders a user listing. IsActive is nil-able so
// asking for inactive users stays distinct from not filtering on status at
// all.
type ListFilters struct {
	Search string
	// Columns narrows individual columns by substring, keyed by FilterColumns.
	// Search casts a wide OR across username, email and role name; these are
	// ANDed on top, so the two answer different questions and compose.
	Columns  map[string]string
	IsActive *bool
	SortKey  string
	SortDesc bool
}

// FilterColumns whitelists the columns a caller may match a substring
// against. Same rule as SortColumns: the column name is interpolated into SQL
// rather than bound, so nothing outside this map may reach the query.
var FilterColumns = map[string]string{
	"username": "u.username",
	"email":    "u.email",
	"role":     "r.name",
}

// SortColumns maps the sort keys the API accepts onto the columns they order
// by. A column name can't be passed to Postgres as a bind parameter, so it is
// interpolated into the query — every value that reaches ORDER BY must come
// from this map and never straight from the request.
var SortColumns = map[string]string{
	"username":   "u.username",
	"email":      "u.email",
	"role":       "r.name",
	"is_active":  "u.is_active",
	"created_at": "u.created_at",
}

const DefaultSortKey = "username"

type Repository interface {
	Count(ctx context.Context, requesterID uuid.UUID, filters ListFilters) (int64, error)
	List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]userdomain.User, error)
	ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]userdomain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (userdomain.User, error)
	// EnsureAccess fails with ErrUserNotFound unless the user exists and the
	// requester may see and act on them. GetByID stays unscoped because the
	// auth flow also loads users through it.
	EnsureAccess(ctx context.Context, id, requesterID uuid.UUID) error
	// EnsureRoleAssignable fails unless the role exists and grants nothing the
	// requester lacks (ErrRoleNotFound / ErrRoleNotAssignable).
	EnsureRoleAssignable(ctx context.Context, roleID, requesterID uuid.UUID) error
	GetPermissions(ctx context.Context, id uuid.UUID) ([]UserPermissionItem, error)
	CountReferences(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput, hashedPassword string) (userdomain.User, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput, hashedPassword *string) (userdomain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ActivityLogger records user create/update/delete events for the audit
// trail. Failures to record are logged but never block the user flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo        Repository
	activityLog ActivityLogger
}

func New(repo Repository, activityLog ActivityLogger) *Service {
	return &Service{repo: repo, activityLog: activityLog}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the user CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "user",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

// List returns every user for roles with full dormitory access, otherwise only
// the users the requester may see: themselves, users they created and users
// sharing a dormitory with them, excluding those with a more privileged role.
func (s *Service) List(ctx context.Context, requesterID uuid.UUID, filters ListFilters, limit, offset int) ([]userdomain.User, int64, error) {
	total, err := s.repo.Count(ctx, requesterID, filters)
	if err != nil {
		return nil, 0, err
	}

	users, err := s.repo.List(ctx, requesterID, filters, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (s *Service) ListActive(ctx context.Context, requesterID uuid.UUID, search string, limit int) ([]userdomain.User, error) {
	return s.repo.ListActive(ctx, requesterID, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (userdomain.User, error) {
	if err := s.repo.EnsureAccess(ctx, id, requesterID); err != nil {
		return userdomain.User{}, err
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetPermissions(ctx context.Context, id, requesterID uuid.UUID) ([]UserPermissionItem, error) {
	if err := s.repo.EnsureAccess(ctx, id, requesterID); err != nil {
		return nil, err
	}
	return s.repo.GetPermissions(ctx, id)
}

func (s *Service) Create(ctx context.Context, requesterID uuid.UUID, input CreateInput, ipAddress string) (userdomain.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	if input.Username == "" || input.Email == "" || strings.TrimSpace(input.Password) == "" {
		return userdomain.User{}, userdomain.ErrRequiredUserData
	}
	if input.RoleID == uuid.Nil {
		return userdomain.User{}, userdomain.ErrRoleNotFound
	}
	if err := s.repo.EnsureRoleAssignable(ctx, input.RoleID, requesterID); err != nil {
		return userdomain.User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return userdomain.User{}, err
	}

	user, err := s.repo.Create(ctx, input, string(hashedPassword))
	if err != nil {
		return userdomain.User{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", user.ID, fmt.Sprintf("Created user: %s", user.Username), ipAddress)
	return user, nil
}

func (s *Service) Update(ctx context.Context, id, requesterID uuid.UUID, input UpdateInput, ipAddress string) (userdomain.User, error) {
	if err := s.repo.EnsureAccess(ctx, id, requesterID); err != nil {
		return userdomain.User{}, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return userdomain.User{}, err
	}
	if user.IsProtected {
		return userdomain.User{}, userdomain.ErrUserProtected
	}

	if input.Username != nil {
		username := strings.TrimSpace(*input.Username)
		if username == "" {
			return userdomain.User{}, userdomain.ErrInvalidUsername
		}
		input.Username = &username
	}
	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		if email == "" {
			return userdomain.User{}, userdomain.ErrInvalidEmail
		}
		input.Email = &email
	}
	if input.RoleID != nil {
		if *input.RoleID == uuid.Nil {
			return userdomain.User{}, userdomain.ErrRoleNotFound
		}
		if err := s.repo.EnsureRoleAssignable(ctx, *input.RoleID, requesterID); err != nil {
			return userdomain.User{}, err
		}
	}

	var hashedPassword *string
	if input.Password != nil {
		password := strings.TrimSpace(*input.Password)
		if password == "" {
			return userdomain.User{}, userdomain.ErrInvalidPassword
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return userdomain.User{}, err
		}
		hashedPasswordValue := string(hashed)
		hashedPassword = &hashedPasswordValue
	}

	updated, err := s.repo.Update(ctx, id, input, hashedPassword)
	if err != nil {
		return userdomain.User{}, err
	}

	s.recordActivity(ctx, input.UpdatedBy, "UPDATE", updated.ID, fmt.Sprintf("Updated user: %s", updated.Username), ipAddress)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	if err := s.repo.EnsureAccess(ctx, id, requesterID); err != nil {
		return err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.IsProtected {
		return userdomain.ErrUserProtected
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, fmt.Sprintf("Deleted user: %s", user.Username), ipAddress)
	return nil
}

func (s *Service) CheckDeletion(ctx context.Context, id, requesterID uuid.UUID) (DeletionCheck, error) {
	if err := s.repo.EnsureAccess(ctx, id, requesterID); err != nil {
		return DeletionCheck{}, err
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	recordCount, err := s.repo.CountReferences(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{
		CanDelete:   !user.IsProtected && recordCount == 0,
		RecordCount: recordCount,
		IsProtected: user.IsProtected,
	}, nil
}
