package usecase

import (
	"context"
	"errors"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type DeleteArticleInput struct {
	ID string
}

type DeleteArticleUseCase struct {
	Repo domain.ArticleDeleter
}

func NewDeleteArticleUseCase(repo domain.ArticleDeleter) DeleteArticleUseCase {
	return DeleteArticleUseCase{Repo: repo}
}

func (uc DeleteArticleUseCase) Execute(ctx context.Context, in DeleteArticleInput) error {
	if uc.Repo == nil {
		return errors.New("delete article: repo is nil")
	}
	if in.ID == "" {
		return errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}
	return uc.Repo.DeleteArticle(ctx, in.ID)
}
