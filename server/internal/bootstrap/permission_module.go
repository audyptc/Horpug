package bootstrap

import "apigofiberhorpug/internal/usecase"

func buildPermissionUseCase(repos repositories) *usecase.PermissionUseCase {
	return usecase.NewPermissionUseCase(repos.perm)
}
