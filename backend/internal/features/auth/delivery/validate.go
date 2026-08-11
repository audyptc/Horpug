package delivery

import (
	"apigofiberhorpug/internal/features/auth/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

func validateLoginRequest(req *domain.LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return apierror.BadRequest("email and password are required")
	}
	return nil
}
