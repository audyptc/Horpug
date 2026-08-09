package delivery

import (
	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/feature/user/domain"
)

func validateCreateUserRequest(req *domain.CreateUserRequest) error {
	if req.FullName == "" || req.Email == "" || req.Password == "" {
		return apierror.BadRequest("full_name, email and password are required")
	}
	return nil
}

func validateAssignRoleRequest(req *domain.AssignRoleRequest) error {
	if req.RoleID == "" {
		return apierror.BadRequest("role_id is required")
	}
	return nil
}
