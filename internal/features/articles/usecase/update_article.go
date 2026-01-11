package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type UpdateArticleInput struct {
	ID       string
	Title    *string
	Content  *string
	Status   *domain.ArticleStatus
	Tags     *[]string
	AutoSave bool
}

type UpdateArticleUseCase struct {
	Repo  domain.ArticleUpdater
	Clock domain.Clock
}

func NewUpdateArticleUseCase(repo domain.ArticleUpdater) UpdateArticleUseCase {
	return UpdateArticleUseCase{Repo: repo, Clock: systemClock{}}
}

func (uc UpdateArticleUseCase) Execute(ctx context.Context, in UpdateArticleInput) (domain.Article, error) {
	if uc.Repo == nil {
		return domain.Article{}, errors.New("update article: repo is nil")
	}
	if uc.Clock == nil {
		uc.Clock = systemClock{}
	}
	if in.ID == "" {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("id is required"))
	}

	status := in.Status
	if in.AutoSave && status == nil {
		ds := domain.ArticleStatusDraft
		status = &ds
	}
	if status != nil && !status.Valid() {
		return domain.Article{}, errors.Join(domain.ErrInvalidArgument, errors.New("invalid status"))
	}

	var normalizedTagsPtr *[]string
	if in.Tags != nil {
		normalized, err := domain.NormalizeTagNames(*in.Tags)
		if err != nil {
			return domain.Article{}, errors.Join(domain.ErrInvalidArgument, err)
		}
		normalizedTagsPtr = &normalized
	}

	now := uc.Clock.Now()
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return uc.Repo.UpdateArticle(ctx, in.ID, domain.UpdateArticleParams{
		Title:      in.Title,
		Content:    in.Content,
		Status:     status,
		Tags:       normalizedTagsPtr,
		UpdatedAt:  now,
		IsAutoSave: in.AutoSave,
	})
}
