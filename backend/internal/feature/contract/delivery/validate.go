package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/contract/domain"
)

func validateCreateContractRequest(req *domain.CreateContractRequest) error {
	if req.TenantID == "" {
		return apierror.BadRequest("tenant_id is required")
	}
	if req.RoomID == "" {
		return apierror.BadRequest("room_id is required")
	}
	if req.StartDate.IsZero() {
		return apierror.BadRequest("start_date is required")
	}
	if req.RentPrice <= 0 {
		return apierror.BadRequest("rent_price must be greater than 0")
	}
	return nil
}
