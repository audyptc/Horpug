package usecase

import (
	"context"
	"math"

	paymentdomain "apihorpug/internal/features/payment/domain"

	"github.com/google/uuid"
)

// ReceiptRepository is the part of the repository the printable receipt needs.
type ReceiptRepository interface {
	GetReceipt(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Receipt, error)
}

// GetReceipt returns the payment as a printable receipt with the invoice's
// remaining balance.
func (s *Service) GetReceipt(ctx context.Context, id, requesterID uuid.UUID) (paymentdomain.Receipt, error) {
	receipt, err := s.repo.GetReceipt(ctx, id, requesterID)
	if err != nil {
		return paymentdomain.Receipt{}, err
	}
	// Rounded to the satang: amounts are NUMERIC(10,2) read into float64.
	receipt.Invoice.Outstanding = math.Max(0, math.Round((receipt.Invoice.TotalAmount-receipt.Invoice.PaidAmount)*100)/100)
	return receipt, nil
}
