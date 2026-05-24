package bootstrap

import "apigofiberhorpug/internal/usecase"

func buildUserUseCase(repos repositories) *usecase.UserUseCase {
	return usecase.NewUserUseCase(repos.user, repos.role)
}
