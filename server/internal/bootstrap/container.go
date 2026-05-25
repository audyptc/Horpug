package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/usecase"
)

type Container struct {
	AuthUC *usecase.AuthUseCase
	UserUC *usecase.UserUseCase
	RoleUC *usecase.RoleUseCase
	PermUC *usecase.PermissionUseCase
	MenuUC *usecase.MenuUseCase
}

func NewContainer(db *database.DB, secretKey string) *Container {
	repos := newRepositories(db)
	return &Container{
		AuthUC: usecase.NewAuthUseCase(repos.user, repos.token, secretKey),
		UserUC: usecase.NewUserUseCase(repos.user, repos.role),
		RoleUC: usecase.NewRoleUseCase(repos.role, repos.perm, repos.menu),
		PermUC: usecase.NewPermissionUseCase(repos.perm),
		MenuUC: usecase.NewMenuUseCase(repos.menu),
	}
}
