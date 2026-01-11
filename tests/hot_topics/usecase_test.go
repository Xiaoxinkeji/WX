
package hottopics_test

import (
	"context"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/usecase"
)

type repoFake struct{ called int }

func (r *repoFake) GetHotTopics(ctx context.Context, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	r.called++
	return []domain.Topic{}, nil
}

func (r *repoFake) RefreshHotTopics(ctx context.Context, source *domain.Source) ([]domain.Topic, error) {
	r.called++
	return []domain.Topic{}, nil
}

func (r *repoFake) SearchHotTopics(ctx context.Context, query string, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	r.called++
	return []domain.Topic{}, nil
}

func TestUseCases_ConstructAndExecute(t *testing.T) {
	repo := &repoFake{}
	fetch := usecase.NewFetchTopicsUseCase(repo)
	_, _ = fetch.Execute(context.Background(), usecase.FetchTopicsInput{})

	refresh := usecase.NewRefreshTopicsUseCase(repo)
	_, _ = refresh.Execute(context.Background(), usecase.RefreshTopicsInput{})

	search := usecase.NewSearchTopicsUseCase(repo)
	_, _ = search.Execute(context.Background(), usecase.SearchTopicsInput{Query: "q"})

	if repo.called != 3 {
		t.Fatalf("expected 3 repo calls, got %d", repo.called)
	}
}
