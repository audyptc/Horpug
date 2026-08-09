package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/auth/domain"
)

func validateLoginRequest(req *domain.LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return apierror.BadRequest("email and password are required")
	}
	return nil
}
