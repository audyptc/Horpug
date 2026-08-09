package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/role/domain"
)

func validateCreateRoleRequest(req *domain.CreateRoleRequest) error {
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}
	return nil
}
