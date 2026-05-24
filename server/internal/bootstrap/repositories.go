package bootstrap

import (
	"apigofiberhorpug/internal/database"
	"apigofiberhorpug/internal/repository"
)

type repositories struct {
	user  *repository.UserRepo
	role  *repository.RoleRepo
	perm  *repository.PermissionRepo
	token *repository.TokenRepo
	menu  *repository.MenuRepo
}

func newRepositories(db *database.DB) repositories {
	return repositories{
		user:  repository.NewUserRepo(db),
		role:  repository.NewRoleRepo(db),
		perm:  repository.NewPermissionRepo(db),
		token: repository.NewTokenRepo(db),
		menu:  repository.NewMenuRepo(db),
	}
}
