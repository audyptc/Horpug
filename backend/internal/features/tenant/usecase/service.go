package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	tenantdomain "apihorpug/internal/features/tenant/domain"

	"github.com/google/uuid"
)

type CreateInput struct {
	FirstName        string
	LastName         string
	Phone            string
	LineID           string
	IDCard           string
	Email            string
	EmergencyContact string
	Note             string
	IsActive         bool
	CreatedBy        *uuid.UUID
}

type UpdateInput struct {
	FirstName        *string
	LastName         *string
	Phone            *string
	LineID           *string
	IDCard           *string
	Email            *string
	EmergencyContact *string
	Note             *string
	IsActive         *bool
	UpdatedBy        *uuid.UUID
}

type DeletionCheck struct {
	CanDelete     bool  `json:"can_delete"`
	ContractCount int64 `json:"contract_count"`
}

type Repository interface {
	Count(ctx context.Context) (int64, error)
	List(ctx context.Context, limit, offset int) ([]tenantdomain.Tenant, error)
	ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error)
	CountContracts(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (tenantdomain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (tenantdomain.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLineUserID(ctx context.Context, id uuid.UUID, lineUserID string) (tenantdomain.Tenant, error)
}

// LineVerifier confirms a LIFF id token was issued by this app's LINE
// channel and returns the LINE userId it belongs to.
type LineVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (string, error)
}

// ActivityLogger records tenant create/update/delete events for the audit
// trail. Failures to record are logged but never block the tenant flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo         Repository
	activityLog  ActivityLogger
	lineVerifier LineVerifier
}

func New(repo Repository, activityLog ActivityLogger, lineVerifier LineVerifier) *Service {
	return &Service{repo: repo, activityLog: activityLog, lineVerifier: lineVerifier}
}

// recordActivity is best-effort: a failure to write the audit trail must
// never fail the tenant CRUD flow itself.
func (s *Service) recordActivity(ctx context.Context, userID *uuid.UUID, action string, entityID uuid.UUID, description, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	_, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      userID,
		Action:      action,
		EntityType:  "tenant",
		EntityID:    &entityID,
		Description: description,
		IPAddress:   ipAddress,
	})
	if err != nil {
		log.Printf("failed to record activity log (action=%s): %v", action, err)
	}
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]tenantdomain.Tenant, int64, error) {
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	tenants, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return tenants, total, nil
}

func (s *Service) ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error) {
	return s.repo.ListActive(ctx, strings.TrimSpace(search), limit)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput, ipAddress string) (tenantdomain.Tenant, error) {
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)
	input.Phone = strings.TrimSpace(input.Phone)
	input.LineID = strings.TrimSpace(input.LineID)
	input.IDCard = strings.TrimSpace(input.IDCard)
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.EmergencyContact = strings.TrimSpace(input.EmergencyContact)
	input.Note = strings.TrimSpace(input.Note)

	if input.FirstName == "" || input.LastName == "" {
		return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
	}

	tenant, err := s.repo.Create(ctx, input)
	if err != nil {
		return tenantdomain.Tenant{}, err
	}

	s.recordActivity(ctx, input.CreatedBy, "CREATE", tenant.ID, fmt.Sprintf("Created tenant: %s %s", tenant.FirstName, tenant.LastName), ipAddress)
	return tenant, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput, ipAddress string) (tenantdomain.Tenant, error) {
	if input.FirstName != nil {
		firstName := strings.TrimSpace(*input.FirstName)
		if firstName == "" {
			return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
		}
		input.FirstName = &firstName
	}
	if input.LastName != nil {
		lastName := strings.TrimSpace(*input.LastName)
		if lastName == "" {
			return tenantdomain.Tenant{}, tenantdomain.ErrRequiredTenantData
		}
		input.LastName = &lastName
	}
	if input.Phone != nil {
		phone := strings.TrimSpace(*input.Phone)
		input.Phone = &phone
	}
	if input.LineID != nil {
		lineID := strings.TrimSpace(*input.LineID)
		input.LineID = &lineID
	}
	if input.IDCard != nil {
		idCard := strings.TrimSpace(*input.IDCard)
		input.IDCard = &idCard
	}
	if input.Email != nil {
		email := strings.TrimSpace(strings.ToLower(*input.Email))
		input.Email = &email
	}
	if input.EmergencyContact != nil {
		emergencyContact := strings.TrimSpace(*input.EmergencyContact)
		input.EmergencyContact = &emergencyContact
	}
	if input.Note != nil {
		note := strings.TrimSpace(*input.Note)
		input.Note = &note
	}

	tenant, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return tenantdomain.Tenant{}, err
	}

	s.recordActivity(ctx, input.UpdatedBy, "UPDATE", tenant.ID, fmt.Sprintf("Updated tenant: %s %s", tenant.FirstName, tenant.LastName), ipAddress)
	return tenant, nil
}

func (s *Service) Delete(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) error {
	tenant, _ := s.repo.GetByID(ctx, id)

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.recordActivity(ctx, &requesterID, "DELETE", id, fmt.Sprintf("Deleted tenant: %s %s", tenant.FirstName, tenant.LastName), ipAddress)
	return nil
}

// LinkLine verifies a LIFF id token (obtained client-side after the tenant
// opens their personal linking link and logs into LINE) and stores the
// resulting LINE userId on the tenant, so future invoices can be pushed to
// them directly through the OA.
func (s *Service) LinkLine(ctx context.Context, id uuid.UUID, idToken string) (tenantdomain.Tenant, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return tenantdomain.Tenant{}, tenantdomain.ErrInvalidLineToken
	}

	lineUserID, err := s.lineVerifier.VerifyIDToken(ctx, idToken)
	if err != nil {
		return tenantdomain.Tenant{}, tenantdomain.ErrInvalidLineToken
	}

	return s.repo.UpdateLineUserID(ctx, id, lineUserID)
}

func (s *Service) CheckDeletion(ctx context.Context, id uuid.UUID) (DeletionCheck, error) {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return DeletionCheck{}, err
	}

	contractCount, err := s.repo.CountContracts(ctx, id)
	if err != nil {
		return DeletionCheck{}, err
	}

	return DeletionCheck{CanDelete: contractCount == 0, ContractCount: contractCount}, nil
}
