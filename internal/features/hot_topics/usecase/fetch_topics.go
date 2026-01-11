
package usecase

import (
	"context"
	"errors"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type FetchTopicsInput struct {
	Source       *domain.Source
	ForceRefresh bool
}

type FetchTopicsUseCase struct {
	Repo domain.Repository
}

func NewFetchTopicsUseCase(repo domain.Repository) FetchTopicsUseCase {
	return FetchTopicsUseCase{Repo: repo}
}

func (uc FetchTopicsUseCase) Execute(ctx context.Context, in FetchTopicsInput) ([]domain.Topic, error) {
	if uc.Repo == nil {
		return nil, errors.New("fetch topics: repo is nil")
	}
	if in.Source != nil && !in.Source.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid source"))
	}
	return uc.Repo.GetHotTopics(ctx, in.Source, in.ForceRefresh)
}
