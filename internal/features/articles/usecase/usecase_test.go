package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
	"github.com/Xiaoxinkeji/WX/internal/features/articles/usecase"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type fixedIDs struct{ id string }

func (g fixedIDs) NewID() (string, error) { return g.id, nil }

type createRepoFake struct{
	called int
	params domain.CreateArticleParams
	ret    domain.Article
	err    error
}

func (f *createRepoFake) CreateArticle(ctx context.Context, params domain.CreateArticleParams) (domain.Article, error) {
	f.called++
	f.params = params
	return f.ret, f.err
}

func TestCreateArticleUseCase_DefaultsAndTagNormalization(t *testing.T) {
	now := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	repo := &createRepoFake{ret: domain.Article{ID: "x"}}
	uc := usecase.CreateArticleUseCase{Repo: repo, Clock: fixedClock{t: now}, IDs: fixedIDs{id: "id-1"}}

	_, err := uc.Execute(context.Background(), usecase.CreateArticleInput{
		Title:   "",
		Content: "",
		Tags:    []string{" Go ", "go", "SQL"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.called != 1 {
		t.Fatalf("expected repo called once, got %d", repo.called)
	}
	if repo.params.ID != "id-1" {
		t.Fatalf("expected id-1, got %q", repo.params.ID)
	}
	if repo.params.Status != domain.ArticleStatusDraft {
		t.Fatalf("expected default draft, got %q", repo.params.Status)
	}
	if !reflect.DeepEqual(repo.params.Tags, []string{"go", "sql"}) {
		t.Fatalf("unexpected tags: %v", repo.params.Tags)
	}
	if !repo.params.CreatedAt.Equal(now) || !repo.params.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected timestamps: %v %v", repo.params.CreatedAt, repo.params.UpdatedAt)
	}
}

func TestCreateArticleUseCase_PublishedRequiresFields(t *testing.T) {
	repo := &createRepoFake{}
	uc := usecase.NewCreateArticleUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.CreateArticleInput{Status: domain.ArticleStatusPublished})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

type updateRepoFake struct{
	called int
	id     string
	params domain.UpdateArticleParams
	ret    domain.Article
	err    error
}

func (f *updateRepoFake) UpdateArticle(ctx context.Context, id string, params domain.UpdateArticleParams) (domain.Article, error) {
	f.called++
	f.id = id
	f.params = params
	return f.ret, f.err
}

func TestUpdateArticleUseCase_AutoSaveForcesDraftWhenUnset(t *testing.T) {
	now := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	repo := &updateRepoFake{ret: domain.Article{ID: "a"}}
	uc := usecase.UpdateArticleUseCase{Repo: repo, Clock: fixedClock{t: now}}

	_, err := uc.Execute(context.Background(), usecase.UpdateArticleInput{
		ID:       "a",
		AutoSave: true,
		Tags:     &[]string{" Go "},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.called != 1 {
		t.Fatalf("expected repo called once, got %d", repo.called)
	}
	if repo.params.Status == nil || *repo.params.Status != domain.ArticleStatusDraft {
		t.Fatalf("expected status draft, got %v", repo.params.Status)
	}
	if repo.params.Tags == nil || !reflect.DeepEqual(*repo.params.Tags, []string{"go"}) {
		t.Fatalf("unexpected tags: %v", repo.params.Tags)
	}
	if !repo.params.UpdatedAt.Equal(now) {
		t.Fatalf("unexpected updated_at: %v", repo.params.UpdatedAt)
	}
	if !repo.params.IsAutoSave {
		t.Fatalf("expected IsAutoSave true")
	}
}

type listRepoFake struct{
	query domain.ListArticlesQuery
	ret   []domain.Article
	err   error
}

func (f *listRepoFake) ListArticles(ctx context.Context, q domain.ListArticlesQuery) ([]domain.Article, error) {
	f.query = q
	return f.ret, f.err
}

func TestListArticlesUseCase_DefaultPagination(t *testing.T) {
	repo := &listRepoFake{}
	uc := usecase.NewListArticlesUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.ListArticlesInput{Limit: 0, Offset: -1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.query.Limit != 20 || repo.query.Offset != 0 {
		t.Fatalf("unexpected pagination: %+v", repo.query)
	}
}

type searchRepoFake struct{
	query domain.SearchArticlesQuery
	ret   []domain.Article
	err   error
}

func (f *searchRepoFake) SearchArticles(ctx context.Context, q domain.SearchArticlesQuery) ([]domain.Article, error) {
	f.query = q
	return f.ret, f.err
}

func TestSearchArticlesUseCase_RequiresQuery(t *testing.T) {
	repo := &searchRepoFake{}
	uc := usecase.NewSearchArticlesUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.SearchArticlesInput{Query: "  "})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

type deleteRepoFake struct{ id string; called int; err error }

func (f *deleteRepoFake) DeleteArticle(ctx context.Context, articleID string) error {
	f.called++
	f.id = articleID
	return f.err
}

func TestDeleteArticleUseCase_RequiresID(t *testing.T) {
	repo := &deleteRepoFake{}
	uc := usecase.NewDeleteArticleUseCase(repo)
	err := uc.Execute(context.Background(), usecase.DeleteArticleInput{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
