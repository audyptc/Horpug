package validator

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
)

func LoginRequest(req *domain.LoginRequest) error {
	if req.Email == "" || req.Password == "" {
		return apierror.BadRequest("email and password are required")
	}
	return nil
}

func RefreshRequest(req *domain.RefreshRequest) error {
	if req.RefreshToken == "" {
		return apierror.BadRequest("refresh token is required")
	}
	return nil
}

func CreateUserRequest(req *domain.CreateUserRequest) error {
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return apierror.BadRequest("full_name, email and password are required")
	}
	return nil
}

func AssignRoleRequest(req *domain.AssignRoleRequest) error {
	if req.RoleID == "" {
		return apierror.BadRequest("role_id is required")
	}
	return nil
}

func CreateRoleRequest(req *domain.CreateRoleRequest) error {
	if req.Name == "" {
		return apierror.BadRequest("name is required")
	}
	return nil
}
