package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/usecase"
)

// Container holds wired application use cases.
type Container struct {
	AuthUC *usecase.AuthUseCase
	UserUC *usecase.UserUseCase
	RoleUC *usecase.RoleUseCase
	PermUC *usecase.PermissionUseCase
	MenuUC *usecase.MenuUseCase
}

// NewContainer builds repositories and use cases in one place.
func NewContainer(db *database.DB, secretKey string) *Container {
	repos := newRepositories(db)

	return &Container{
		AuthUC: buildAuthUseCase(repos, secretKey),
		UserUC: buildUserUseCase(repos),
		RoleUC: buildRoleUseCase(repos),
		PermUC: buildPermissionUseCase(repos),
		MenuUC: buildMenuUseCase(repos),
	}
}
