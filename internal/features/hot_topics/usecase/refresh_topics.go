
package usecase

import (
	"context"
	"errors"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type RefreshTopicsInput struct {
	Source *domain.Source
}

type RefreshTopicsUseCase struct {
	Repo domain.Repository
}

func NewRefreshTopicsUseCase(repo domain.Repository) RefreshTopicsUseCase {
	return RefreshTopicsUseCase{Repo: repo}
}

func (uc RefreshTopicsUseCase) Execute(ctx context.Context, in RefreshTopicsInput) ([]domain.Topic, error) {
	if uc.Repo == nil {
		return nil, errors.New("refresh topics: repo is nil")
	}
	if in.Source != nil && !in.Source.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid source"))
	}
	return uc.Repo.RefreshHotTopics(ctx, in.Source)
}
