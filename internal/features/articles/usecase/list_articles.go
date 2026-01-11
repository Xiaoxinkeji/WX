package usecase

import (
	"context"
	"errors"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type ListArticlesInput struct {
	Status *domain.ArticleStatus
	Tag    *string
	Limit  int
	Offset int
}

type ListArticlesUseCase struct {
	Repo domain.ArticleLister
}

func NewListArticlesUseCase(repo domain.ArticleLister) ListArticlesUseCase {
	return ListArticlesUseCase{Repo: repo}
}

func (uc ListArticlesUseCase) Execute(ctx context.Context, in ListArticlesInput) ([]domain.Article, error) {
	if uc.Repo == nil {
		return nil, errors.New("list articles: repo is nil")
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

	return uc.Repo.ListArticles(ctx, domain.ListArticlesQuery{
		Status: in.Status,
		Tag:    in.Tag,
		Limit:  limit,
		Offset: offset,
	})
}
