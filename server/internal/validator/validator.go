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

func CreateRoomRequest(req *domain.CreateRoomRequest) error {
	if req.RoomNumber == "" {
		return apierror.BadRequest("room_number is required")
	}
	if req.Floor <= 0 {
		return apierror.BadRequest("floor must be greater than 0")
	}
	if req.RentPrice <= 0 {
		return apierror.BadRequest("rent_price must be greater than 0")
	}
	return nil
}
