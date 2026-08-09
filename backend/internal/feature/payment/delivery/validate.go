package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/payment/domain"
)

func validatePaymentSplits(splits []domain.PaymentSplitRequest) error {
	validMethods := map[domain.PaymentMethod]bool{
		domain.PaymentMethodCash:     true,
		domain.PaymentMethodTransfer: true,
		domain.PaymentMethodQR:       true,
	}
	if len(splits) == 0 {
		return apierror.BadRequest("splits must have at least one entry")
	}
	if len(splits) > 3 {
		return apierror.BadRequest("splits must have at most 3 entries")
	}
	seen := map[domain.PaymentMethod]bool{}
	for _, s := range splits {
		if !validMethods[s.Method] {
			return apierror.BadRequest("split method must be one of: cash, transfer, qr")
		}
		if s.Amount <= 0 {
			return apierror.BadRequest("each split amount must be greater than 0")
		}
		if seen[s.Method] {
			return apierror.BadRequest("duplicate split method: " + string(s.Method))
		}
		seen[s.Method] = true
	}
	return nil
}

func validateCreatePaymentRequest(req *domain.CreatePaymentRequest) error {
	if req.BillID == "" {
		return apierror.BadRequest("bill_id is required")
	}
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return validatePaymentSplits(req.Splits)
}

func validateUpdatePaymentRequest(req *domain.UpdatePaymentRequest) error {
	if req.PaymentDate.IsZero() {
		return apierror.BadRequest("payment_date is required")
	}
	return validatePaymentSplits(req.Splits)
}
