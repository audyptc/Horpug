package usecase

import (
	"context"
	"strings"

	"apigofiberhorpug/internal/delivery/http/apierror"
	"apigofiberhorpug/internal/domain"
)

type searchRepository interface {
	Search(ctx context.Context, q string) (*domain.SearchResults, error)
}

type SearchUseCase struct {
	repo searchRepository
}

func NewSearchUseCase(repo searchRepository) *SearchUseCase {
	return &SearchUseCase{repo: repo}
}

func (uc *SearchUseCase) Search(ctx context.Context, q string) (*domain.SearchResults, error) {
	q = strings.TrimSpace(q)
	if len([]rune(q)) < 2 {
		return nil, apierror.BadRequest("query must be at least 2 characters")
	}
	return uc.repo.Search(ctx, q)
}
