package bootstrap

import "apigofiberhorpug/internal/usecase"

func buildMenuUseCase(repos repositories) *usecase.MenuUseCase {
	return usecase.NewMenuUseCase(repos.menu)
}
