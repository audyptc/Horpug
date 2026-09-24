package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	activitylogdomain "apihorpug/internal/features/activitylog/domain"
	activitylogusecase "apihorpug/internal/features/activitylog/usecase"
	moveoutdomain "apihorpug/internal/features/moveout/domain"

	"github.com/google/uuid"
)

type Repository interface {
	LoadForPreview(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time) (moveoutdomain.Settlement, Inputs, error)
	Settle(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time, note string,
		build func(base moveoutdomain.Settlement, in Inputs) (moveoutdomain.Settlement, error)) (uuid.UUID, error)
	GetByID(ctx context.Context, id, requesterID uuid.UUID) (moveoutdomain.MoveOut, error)
	ListPresets(ctx context.Context, dormitoryID, requesterID uuid.UUID) ([]moveoutdomain.Preset, error)
	PresetsForDormitory(ctx context.Context, dormitoryID uuid.UUID) ([]moveoutdomain.Preset, error)
	ReplacePresets(ctx context.Context, dormitoryID, requesterID uuid.UUID, presets []moveoutdomain.Preset) ([]moveoutdomain.Preset, error)
}

// ActivityLogger records confirmed move-outs for the audit trail. Failures
// are logged but never block the move-out.
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

// Preview works out the settlement for moving out on moveOutDate with the
// given deductions, without changing anything, along with the dormitory's
// ready-made deductions.
func (s *Service) Preview(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time, manual []ManualItem) (moveoutdomain.Preview, error) {
	if moveOutDate.IsZero() {
		return moveoutdomain.Preview{}, moveoutdomain.ErrRequiredDate
	}
	moveOutDate = dateOnly(moveOutDate)

	base, in, err := s.repo.LoadForPreview(ctx, contractID, requesterID, moveOutDate)
	if err != nil {
		return moveoutdomain.Preview{}, err
	}
	base.MoveOutDate = moveOutDate
	settlement, err := BuildSettlement(base, in, manual)
	if err != nil {
		return moveoutdomain.Preview{}, err
	}
	presets, err := s.repo.PresetsForDormitory(ctx, settlement.DormitoryID)
	if err != nil {
		return moveoutdomain.Preview{}, err
	}
	return moveoutdomain.Preview{Settlement: settlement, Presets: presets}, nil
}

// Confirm settles the move-out: the deposit pays the contract's unpaid
// invoices, the contract is terminated and the room freed. The settlement is
// rebuilt from the data as it stands inside the transaction, so it matches
// what is recorded even if something changed since the preview.
func (s *Service) Confirm(ctx context.Context, contractID, requesterID uuid.UUID, moveOutDate time.Time, note string, manual []ManualItem, ipAddress string) (moveoutdomain.MoveOut, error) {
	if moveOutDate.IsZero() {
		return moveoutdomain.MoveOut{}, moveoutdomain.ErrRequiredDate
	}
	moveOutDate = dateOnly(moveOutDate)
	manual, err := normalizeManualItems(manual)
	if err != nil {
		return moveoutdomain.MoveOut{}, err
	}

	id, err := s.repo.Settle(ctx, contractID, requesterID, moveOutDate, strings.TrimSpace(note),
		func(base moveoutdomain.Settlement, in Inputs) (moveoutdomain.Settlement, error) {
			base.MoveOutDate = moveOutDate
			return BuildSettlement(base, in, manual)
		})
	if err != nil {
		return moveoutdomain.MoveOut{}, err
	}

	moveOut, err := s.repo.GetByID(ctx, id, requesterID)
	if err != nil {
		return moveoutdomain.MoveOut{}, err
	}

	s.recordActivity(ctx, requesterID, moveOut, ipAddress)
	return moveOut, nil
}

func (s *Service) GetByID(ctx context.Context, id, requesterID uuid.UUID) (moveoutdomain.MoveOut, error) {
	return s.repo.GetByID(ctx, id, requesterID)
}

func (s *Service) ListPresets(ctx context.Context, dormitoryID, requesterID uuid.UUID) ([]moveoutdomain.Preset, error) {
	return s.repo.ListPresets(ctx, dormitoryID, requesterID)
}

// ReplacePresets validates and stores a dormitory's ready-made deductions.
func (s *Service) ReplacePresets(ctx context.Context, dormitoryID, requesterID uuid.UUID, presets []moveoutdomain.Preset) ([]moveoutdomain.Preset, error) {
	clean := make([]moveoutdomain.Preset, 0, len(presets))
	for _, p := range presets {
		p.Name = strings.TrimSpace(p.Name)
		p.Amount = round2(p.Amount)
		if p.Name == "" || p.Amount <= 0 {
			return nil, moveoutdomain.ErrInvalidPreset
		}
		clean = append(clean, p)
	}
	return s.repo.ReplacePresets(ctx, dormitoryID, requesterID, clean)
}

func (s *Service) recordActivity(ctx context.Context, userID uuid.UUID, m moveoutdomain.MoveOut, ipAddress string) {
	if s.activityLog == nil {
		return
	}
	dormitoryID := m.DormitoryID
	description := fmt.Sprintf("Moved out: %s - room %s on %s, deposit %.2f, deductions %.2f, refund %.2f, due %.2f",
		m.TenantName, m.RoomNumber, m.MoveOutDate.Format("2006-01-02"), m.Deposit, m.TotalDeductions, m.RefundAmount, m.AmountDue)
	if _, err := s.activityLog.Create(ctx, activitylogusecase.CreateInput{
		UserID:      &userID,
		Action:      "MOVE_OUT",
		EntityType:  "contract",
		EntityID:    &m.ContractID,
		DormitoryID: &dormitoryID,
		Description: description,
		IPAddress:   ipAddress,
	}); err != nil {
		log.Printf("failed to record activity log (action=MOVE_OUT): %v", err)
	}
}
