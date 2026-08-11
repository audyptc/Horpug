package delivery

import (
	"apigofiberhorpug/internal/feature/role/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateRoleRequest(req *domain.CreateRoleRequest) error {
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}
	return nil
}
