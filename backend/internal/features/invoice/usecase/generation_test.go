package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	invoicedomain "apihorpug/internal/features/invoice/domain"

	"github.com/google/uuid"
)

// fakeGenerationRepo implements only what Generate touches; the embedded nil
// Repository panics if anything else is called.
type fakeGenerationRepo struct {
	Repository
	candidates []invoicedomain.GenerationCandidate
	createErr  map[uuid.UUID]error
	created    []uuid.UUID
}

func (f *fakeGenerationRepo) ListGenerationCandidates(context.Context, uuid.UUID, uuid.UUID, int, int) ([]invoicedomain.GenerationCandidate, error) {
	return f.candidates, nil
}

func (f *fakeGenerationRepo) Create(_ context.Context, input CreateInput) (invoicedomain.Invoice, error) {
	if err := f.createErr[input.ContractID]; err != nil {
		return invoicedomain.Invoice{}, err
	}
	f.created = append(f.created, input.ContractID)
	return invoicedomain.Invoice{ID: uuid.New(), ContractID: input.ContractID, TotalAmount: 3500}, nil
}

func validGenerateInput() GenerateInput {
	return GenerateInput{
		DormitoryID: uuid.New(),
		PeriodYear:  2026,
		PeriodMonth: 9,
		IssueDate:   time.Date(2026, time.September, 25, 0, 0, 0, 0, time.UTC),
		DueDate:     time.Date(2026, time.October, 5, 0, 0, 0, 0, time.UTC),
		CreatedBy:   uuid.New(),
	}
}

func TestGenerateBillsOnlyUninvoicedContracts(t *testing.T) {
	billed, fresh, failing, racing := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo := &fakeGenerationRepo{
		candidates: []invoicedomain.GenerationCandidate{
			{ContractID: billed, AlreadyInvoiced: true},
			{ContractID: fresh, RoomNumber: "101"},
			{ContractID: failing, RoomNumber: "102"},
			{ContractID: racing, RoomNumber: "103"},
		},
		createErr: map[uuid.UUID]error{
			failing: errors.New("boom"),
			racing:  invoicedomain.ErrInvoiceExists,
		},
	}
	svc := New(repo, nil, nil)

	result, err := svc.Generate(context.Background(), validGenerateInput(), "")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(result.Created) != 1 || result.Created[0].ContractID != fresh || result.Created[0].RoomNumber != "101" {
		t.Errorf("Created = %+v, want only contract %s", result.Created, fresh)
	}
	if result.Skipped != 2 {
		t.Errorf("Skipped = %d, want 2 (already invoiced + created concurrently)", result.Skipped)
	}
	if len(result.Failed) != 1 || result.Failed[0].ContractID != failing {
		t.Errorf("Failed = %+v, want only contract %s", result.Failed, failing)
	}
}

func TestGenerateRespectsSelectedContracts(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	repo := &fakeGenerationRepo{candidates: []invoicedomain.GenerationCandidate{{ContractID: a}, {ContractID: b}}}
	svc := New(repo, nil, nil)

	input := validGenerateInput()
	input.ContractIDs = []uuid.UUID{b}
	if _, err := svc.Generate(context.Background(), input, ""); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(repo.created) != 1 || repo.created[0] != b {
		t.Errorf("created = %v, want only %s", repo.created, b)
	}
}

func TestGenerateValidatesInput(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*GenerateInput)
		want   error
	}{
		{"missing dormitory", func(in *GenerateInput) { in.DormitoryID = uuid.Nil }, invoicedomain.ErrRequiredGenerateData},
		{"bad month", func(in *GenerateInput) { in.PeriodMonth = 13 }, invoicedomain.ErrInvalidInvoicePeriod},
		{"due before issue", func(in *GenerateInput) { in.DueDate = in.IssueDate.AddDate(0, 0, -1) }, invoicedomain.ErrInvalidInvoiceDates},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := New(&fakeGenerationRepo{}, nil, nil)
			input := validGenerateInput()
			tc.mutate(&input)
			if _, err := svc.Generate(context.Background(), input, ""); !errors.Is(err, tc.want) {
				t.Fatalf("Generate() error = %v, want %v", err, tc.want)
			}
		})
	}
}
