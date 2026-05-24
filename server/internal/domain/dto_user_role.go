package domain

type CreateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	IsActive *bool  `json:"is_active"`
}

type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []string `json:"permission_ids"`
}
