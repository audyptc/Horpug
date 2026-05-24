package domain

type CreateUserRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	FullName string `json:"full_name"`
	Password string `json:"password"`
	IsActive *bool  `json:"is_active"`
}

type AssignRoleRequest struct {
	RoleID string `json:"role_id"`
}

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleMenuPermissionItem struct {
	MenuID        string   `json:"menu_id"`
	PermissionIDs []string `json:"permission_ids"`
}

type AssignPermissionsRequest struct {
	Items []RoleMenuPermissionItem `json:"items"`
}
