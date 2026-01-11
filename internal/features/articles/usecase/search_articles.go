package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type SearchArticlesInput struct {
	Query  string
	Status *domain.ArticleStatus
	Tag    *string
	Limit  int
	Offset int
}

type SearchArticlesUseCase struct {
	Repo domain.ArticleSearcher
}

func NewSearchArticlesUseCase(repo domain.ArticleSearcher) SearchArticlesUseCase {
	return SearchArticlesUseCase{Repo: repo}
}

func (uc SearchArticlesUseCase) Execute(ctx context.Context, in SearchArticlesInput) ([]domain.Article, error) {
	if uc.Repo == nil {
		return nil, errors.New("search articles: repo is nil")
	}
	q := strings.TrimSpace(in.Query)
	if q == "" {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("query is required"))
	}
	if in.Status != nil && !in.Status.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid status"))
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := in.Offset
	if offset < 0 {
		offset = 0
	}

	return uc.Repo.SearchArticles(ctx, domain.SearchArticlesQuery{
		Query:  q,
		Status: in.Status,
		Tag:    in.Tag,
		Limit:  limit,
		Offset: offset,
	})
}
