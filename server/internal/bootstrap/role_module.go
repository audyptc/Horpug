package bootstrap

import "apigofiberhorpug/internal/usecase"

func buildRoleUseCase(repos repositories) *usecase.RoleUseCase {
	return usecase.NewRoleUseCase(repos.role, repos.perm, repos.menu)
}
