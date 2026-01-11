
package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type SearchTopicsInput struct {
	Query        string
	Source       *domain.Source
	ForceRefresh bool
}

type SearchTopicsUseCase struct {
	Repo domain.Repository
}

func NewSearchTopicsUseCase(repo domain.Repository) SearchTopicsUseCase {
	return SearchTopicsUseCase{Repo: repo}
}

func (uc SearchTopicsUseCase) Execute(ctx context.Context, in SearchTopicsInput) ([]domain.Topic, error) {
	if uc.Repo == nil {
		return nil, errors.New("search topics: repo is nil")
	}
	if in.Source != nil && !in.Source.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid source"))
	}
	query := strings.TrimSpace(in.Query)
	return uc.Repo.SearchHotTopics(ctx, query, in.Source, in.ForceRefresh)
}
