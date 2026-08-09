package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/bill/domain"
)

func validateCreateBillRequest(req *domain.CreateBillRequest) error {
	if req.ContractID == "" {
		return apierror.BadRequest("contract_id is required")
	}
	if req.BillingMonth.IsZero() {
		return apierror.BadRequest("billing_month is required")
	}
	if req.RentAmount < 0 {
		return apierror.BadRequest("rent_amount must be >= 0")
	}
	for _, item := range req.OtherItems {
		if item.Amount < 0 {
			return apierror.BadRequest("other item amount must be >= 0")
		}
	}
	return nil
}
