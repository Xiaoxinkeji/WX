
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type repoFake struct {
	calledGet     int
	calledRefresh int
	calledSearch  int

	lastSource *domain.Source
	lastQuery  string
	lastForce  bool

	topics []domain.Topic
	err    error
}

func (r *repoFake) GetHotTopics(ctx context.Context, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	r.calledGet++
	r.lastSource = source
	r.lastForce = forceRefresh
	return r.topics, r.err
}

func (r *repoFake) RefreshHotTopics(ctx context.Context, source *domain.Source) ([]domain.Topic, error) {
	r.calledRefresh++
	r.lastSource = source
	return r.topics, r.err
}

func (r *repoFake) SearchHotTopics(ctx context.Context, query string, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	r.calledSearch++
	r.lastQuery = query
	r.lastSource = source
	r.lastForce = forceRefresh
	return r.topics, r.err
}

func TestFetchTopicsUseCase_ValidatesRepoAndSource(t *testing.T) {
	uc := FetchTopicsUseCase{}
	_, err := uc.Execute(context.Background(), FetchTopicsInput{})
	if err == nil {
		t.Fatalf("expected error")
	}

	repo := &repoFake{}
	uc = NewFetchTopicsUseCase(repo)
	bad := domain.Source("bad")
	_, err = uc.Execute(context.Background(), FetchTopicsInput{Source: &bad})
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestRefreshAndSearchUseCases_CallRepo(t *testing.T) {
	repo := &repoFake{topics: []domain.Topic{{ID: "x"}}}

	s := domain.SourceWeibo

	refresh := NewRefreshTopicsUseCase(repo)
	_, _ = refresh.Execute(context.Background(), RefreshTopicsInput{Source: &s})
	if repo.calledRefresh != 1 {
		t.Fatalf("expected refresh called")
	}

	search := NewSearchTopicsUseCase(repo)
	_, _ = search.Execute(context.Background(), SearchTopicsInput{Query: "  Q  ", Source: &s, ForceRefresh: true})
	if repo.calledSearch != 1 {
		t.Fatalf("expected search called")
	}
	if repo.lastQuery != "Q" {
		t.Fatalf("expected trimmed query, got %q", repo.lastQuery)
	}
	if !repo.lastForce {
		t.Fatalf("expected force refresh true")
	}
}
