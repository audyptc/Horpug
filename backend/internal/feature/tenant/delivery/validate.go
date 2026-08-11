package delivery

import (
	"apigofiberhorpug/internal/feature/tenant/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateTenantRequest(req *domain.CreateTenantRequest) error {
	if req.FirstName == "" || req.LastName == "" {
		return apierror.BadRequest("first_name and last_name are required")
	}
	if req.IDCard == "" {
		return apierror.BadRequest("id_card is required")
	}
	return nil
}
