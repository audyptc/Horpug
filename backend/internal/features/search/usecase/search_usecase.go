package usecase

import (
	"context"
	"strings"

	"apigofiberhorpug/internal/features/search/domain"
	"apigofiberhorpug/internal/shared/http/apierror"
)

type SearchUseCase struct {
	repo domain.SearchRepository
}

func NewSearchUseCase(repo domain.SearchRepository) *SearchUseCase {
	return &SearchUseCase{repo: repo}
}

func (uc *SearchUseCase) Search(ctx context.Context, dormitoryID, q string) (*domain.SearchResults, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) < 2 {
		return nil, apierror.BadRequest("query must be at least 2 characters")
	}
	return uc.repo.Search(ctx, dormitoryID, q)
}
