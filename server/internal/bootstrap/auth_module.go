package bootstrap

import "apigofiberhorpug/internal/usecase"

func buildAuthUseCase(repos repositories, secretKey string) *usecase.AuthUseCase {
	return usecase.NewAuthUseCase(repos.user, repos.token, secretKey)
}
