package usecase

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

type GenerateInput struct {
	DormitoryID uuid.UUID
	PeriodYear  int
	PeriodMonth int
	IssueDate   time.Time
	DueDate     time.Time
	Note        string
	// ContractIDs limits generation to these contracts; empty means every
	// candidate not yet invoiced for the period.
	ContractIDs []uuid.UUID
	CreatedBy   uuid.UUID
}

// GenerationRepository is the part of the repository bulk generation needs.
type GenerationRepository interface {
	ListGenerationCandidates(ctx context.Context, requesterID, dormitoryID uuid.UUID, periodYear, periodMonth int) ([]invoicedomain.GenerationCandidate, error)
}

func validatePeriod(periodYear, periodMonth int) error {
	if periodYear <= 0 {
		return invoicedomain.ErrRequiredGenerateData
	}
	if periodMonth < 1 || periodMonth > 12 {
		return invoicedomain.ErrInvalidInvoicePeriod
	}
	return nil
}

// PreviewGeneration lists the contracts a bulk generation for the period would
// consider, so staff can spot missing meter readings before billing.
func (s *Service) PreviewGeneration(ctx context.Context, requesterID, dormitoryID uuid.UUID, periodYear, periodMonth int) ([]invoicedomain.GenerationCandidate, error) {
	if dormitoryID == uuid.Nil {
		return nil, invoicedomain.ErrRequiredGenerateData
	}
	if err := validatePeriod(periodYear, periodMonth); err != nil {
		return nil, err
	}
	return s.repo.ListGenerationCandidates(ctx, requesterID, dormitoryID, periodYear, periodMonth)
}

// Generate creates the period's invoice for every candidate contract that
// doesn't have one yet, through the same Create path as a single invoice (so
// each gets its rent and meter items and an activity-log entry).
func (s *Service) Generate(ctx context.Context, input GenerateInput, ipAddress string) (invoicedomain.GenerationResult, error) {
	input.Note = strings.TrimSpace(input.Note)
	if input.DormitoryID == uuid.Nil || input.IssueDate.IsZero() || input.DueDate.IsZero() {
		return invoicedomain.GenerationResult{}, invoicedomain.ErrRequiredGenerateData
	}
	if err := validatePeriod(input.PeriodYear, input.PeriodMonth); err != nil {
		return invoicedomain.GenerationResult{}, err
	}
	if input.DueDate.Before(input.IssueDate) {
		return invoicedomain.GenerationResult{}, invoicedomain.ErrInvalidInvoiceDates
	}

	candidates, err := s.repo.ListGenerationCandidates(ctx, input.CreatedBy, input.DormitoryID, input.PeriodYear, input.PeriodMonth)
	if err != nil {
		return invoicedomain.GenerationResult{}, err
	}

	selected := make(map[uuid.UUID]struct{}, len(input.ContractIDs))
	for _, id := range input.ContractIDs {
		selected[id] = struct{}{}
	}

	result := invoicedomain.GenerationResult{
		Created: make([]invoicedomain.GeneratedInvoice, 0),
		Failed:  make([]invoicedomain.GenerationFailure, 0),
	}
	createdBy := input.CreatedBy
	for _, candidate := range candidates {
		if len(selected) > 0 {
			if _, ok := selected[candidate.ContractID]; !ok {
				continue
			}
		}
		if candidate.AlreadyInvoiced {
			result.Skipped++
			continue
		}

		invoice, err := s.Create(ctx, CreateInput{
			ContractID:  candidate.ContractID,
			PeriodYear:  input.PeriodYear,
			PeriodMonth: input.PeriodMonth,
			IssueDate:   input.IssueDate,
			DueDate:     input.DueDate,
			Note:        input.Note,
			CreatedBy:   &createdBy,
		}, ipAddress)
		if err != nil {
			// Created in the meantime (e.g. a double submit): nothing to do.
			if errors.Is(err, invoicedomain.ErrInvoiceExists) {
				result.Skipped++
				continue
			}
			log.Printf("bulk invoice generation failed (contract=%s): %v", candidate.ContractID, err)
			result.Failed = append(result.Failed, invoicedomain.GenerationFailure{
				ContractID: candidate.ContractID,
				TenantName: candidate.TenantName,
				RoomNumber: candidate.RoomNumber,
				Error:      err.Error(),
			})
			continue
		}

		result.Created = append(result.Created, invoicedomain.GeneratedInvoice{
			InvoiceID:   invoice.ID,
			ContractID:  candidate.ContractID,
			TenantName:  candidate.TenantName,
			RoomNumber:  candidate.RoomNumber,
			TotalAmount: invoice.TotalAmount,
		})
	}

	return result, nil
}

// OverdueMarker flips invoice statuses by due date; see Repository.MarkOverdue.
type OverdueMarker interface {
	MarkOverdue(ctx context.Context) (markedOverdue, revertedUnpaid int64, err error)
}

// sweepOverdue marks overdue invoices once (see RunInvoiceJobs). Without it an
// invoice stays "unpaid" past its due date unless someone changes it by hand.
func sweepOverdue(ctx context.Context, marker OverdueMarker) {
	sweepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	overdue, reverted, err := marker.MarkOverdue(sweepCtx)
	if err != nil {
		log.Printf("overdue invoice sweep failed: %v", err)
		return
	}
	if overdue > 0 || reverted > 0 {
		log.Printf("overdue invoice sweep: %d marked overdue, %d back to unpaid", overdue, reverted)
	}
}
