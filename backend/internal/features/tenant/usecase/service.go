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

// ListFilter narrows and orders a tenant listing. The nil-able flags mean "no
// preference" rather than false, so a caller can ask for inactive tenants
// without that being confused with not filtering on status at all.
type ListFilter struct {
	Search     string
	IsActive   *bool
	LineLinked *bool
	SortKey    string
	SortDesc   bool
}

// SortColumns maps the sort keys the API accepts onto the columns they order
// by. A column name can't be passed to Postgres as a bind parameter, so it is
// interpolated into the query — every value that reaches ORDER BY must come
// from this map and never straight from the request.
var SortColumns = map[string]string{
	"first_name": "first_name",
	"last_name":  "last_name",
	"phone":      "phone",
	"line_id":    "line_id",
	"id_card":    "id_card",
	"email":      "email",
	"is_active":  "is_active",
	"created_at": "created_at",
}

const DefaultSortKey = "created_at"

type Repository interface {
	Count(ctx context.Context, filter ListFilter) (int64, error)
	List(ctx context.Context, filter ListFilter, limit, offset int) ([]tenantdomain.Tenant, error)
	ListActive(ctx context.Context, search string, limit int) ([]tenantdomain.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error)
	CountContracts(ctx context.Context, id uuid.UUID) (int64, error)
	Create(ctx context.Context, input CreateInput) (tenantdomain.Tenant, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (tenantdomain.Tenant, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLineUserID(ctx context.Context, id uuid.UUID, lineUserID string) (tenantdomain.Tenant, error)
	UnlinkLine(ctx context.Context, id uuid.UUID) (tenantdomain.Tenant, error)
}

// LineClient covers the parts of LINE's platform the tenant flow needs:
// confirming a LIFF id token (and the LINE userId behind it), identifying the
// Official Account, and checking whether a user can actually be reached.
type LineClient interface {
	VerifyIDToken(ctx context.Context, idToken string) (string, error)
	GetBotInfo(ctx context.Context) (basicID, displayName string, err error)
	IsFriend(ctx context.Context, lineUserID string) (bool, error)
}

// LineOAInfo identifies the dormitory's LINE Official Account, so tenants can
// be pointed at it to add it as a friend — a prerequisite for receiving any
// pushed invoice, and separate from having linked their account.
type LineOAInfo struct {
	BasicID      string `json:"basic_id"`
	DisplayName  string `json:"display_name"`
	AddFriendURL string `json:"add_friend_url"`
}

// LineStatus reports whether a tenant can actually be sent an invoice over
// LINE. Both conditions are required and are independent: linking gives us a
// userId to push to, and friendship is what LINE requires before it will
// deliver anything to that userId.
type LineStatus struct {
	Linked   bool `json:"linked"`
	IsFriend bool `json:"is_friend"`
}

// ActivityLogger records tenant create/update/delete events for the audit
// trail. Failures to record are logged but never block the tenant flow.
type ActivityLogger interface {
	Create(ctx context.Context, input activitylogusecase.CreateInput) (activitylogdomain.ActivityLog, error)
}

type Service struct {
	repo        Repository
	activityLog ActivityLogger
	lineClient  LineClient
}

func New(repo Repository, activityLog ActivityLogger, lineClient LineClient) *Service {
	return &Service{repo: repo, activityLog: activityLog, lineClient: lineClient}
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

func (s *Service) List(ctx context.Context, filter ListFilter, limit, offset int) ([]tenantdomain.Tenant, int64, error) {
	filter.Search = strings.TrimSpace(filter.Search)
	if _, ok := SortColumns[filter.SortKey]; !ok {
		filter.SortKey = DefaultSortKey
	}

	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	tenants, err := s.repo.List(ctx, filter, limit, offset)
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

	lineUserID, err := s.lineClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return tenantdomain.Tenant{}, tenantdomain.ErrInvalidLineToken
	}

	return s.repo.UpdateLineUserID(ctx, id, lineUserID)
}

// LineOAInfo reports the dormitory's LINE Official Account and the URL that
// adds it as a friend.
func (s *Service) LineOAInfo(ctx context.Context) (LineOAInfo, error) {
	basicID, displayName, err := s.lineClient.GetBotInfo(ctx)
	if err != nil {
		return LineOAInfo{}, err
	}

	return LineOAInfo{
		BasicID:      basicID,
		DisplayName:  displayName,
		AddFriendURL: "https://line.me/R/ti/p/" + basicID,
	}, nil
}

// LineStatus reports whether a tenant is reachable by a pushed invoice. A
// tenant who never linked has no userId to check, so friendship is reported
// as false without calling LINE.
func (s *Service) LineStatus(ctx context.Context, id uuid.UUID) (LineStatus, error) {
	tenant, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return LineStatus{}, err
	}
	if tenant.LineUserID == "" {
		return LineStatus{}, nil
	}

	isFriend, err := s.lineClient.IsFriend(ctx, tenant.LineUserID)
	if err != nil {
		return LineStatus{}, err
	}

	return LineStatus{Linked: true, IsFriend: isFriend}, nil
}

// UnlinkLine clears a tenant's stored LINE userId, e.g. because the wrong
// LINE account was linked, so their linking link can be used again to link
// the correct one.
func (s *Service) UnlinkLine(ctx context.Context, id, requesterID uuid.UUID, ipAddress string) (tenantdomain.Tenant, error) {
	tenant, err := s.repo.UnlinkLine(ctx, id)
	if err != nil {
		return tenantdomain.Tenant{}, err
	}

	s.recordActivity(ctx, &requesterID, "UPDATE", tenant.ID, fmt.Sprintf("Unlinked LINE account for tenant: %s %s", tenant.FirstName, tenant.LastName), ipAddress)
	return tenant, nil
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
