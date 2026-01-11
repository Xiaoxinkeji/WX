package articles_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/data"
	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

func TestSearch_FilterByTagAndStatus(t *testing.T) {
	ctx := context.Background()
	name := fmt.Sprintf("file:articles_search_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	repo, err := data.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	now := time.Now().UTC()
	_, _ = repo.CreateArticle(ctx, domain.CreateArticleParams{ID: "a1", Title: "Alpha", Content: "Golang and SQLite", Status: domain.ArticleStatusPublished, Tags: []string{"go"}, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateArticle(ctx, domain.CreateArticleParams{ID: "a2", Title: "Beta", Content: "SQLite tips", Status: domain.ArticleStatusDraft, Tags: []string{"sql"}, CreatedAt: now, UpdatedAt: now})

	tag := "go"
	status := domain.ArticleStatusPublished
	got, err := repo.SearchArticles(ctx, domain.SearchArticlesQuery{Query: "SQLite", Tag: &tag, Status: &status, Limit: 10})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 || got[0].ID != "a1" {
		t.Fatalf("unexpected results: %+v", got)
	}
}
