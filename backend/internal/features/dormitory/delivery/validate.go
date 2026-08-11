package delivery

import (
	"apigofiberhorpug/internal/features/dormitory/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateCreateDormitoryRequest(req *domain.CreateDormitoryRequest) error {
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}
	return nil
}
