package data_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Xiaoxinkeji/WX/internal/features/articles/data"
	"github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("file:articles_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSQLiteRepository_CRUD_Versions_Tags_Search(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo, err := data.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	createdAt := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	article, err := repo.CreateArticle(ctx, domain.CreateArticleParams{
		ID:        "a1",
		Title:     "Hello Go",
		Content:   "Full text search with SQLite",
		Status:    domain.ArticleStatusDraft,
		Tags:      []string{"Go", "SQLite"},
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if article.ID != "a1" || article.CurrentVersion != 1 {
		t.Fatalf("unexpected article: %+v", article)
	}
	if len(article.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(article.Tags))
	}

	got, err := repo.GetArticle(ctx, "a1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Title != "Hello Go" {
		t.Fatalf("unexpected title: %q", got.Title)
	}

	results, err := repo.SearchArticles(ctx, domain.SearchArticlesQuery{Query: "Hello"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 1 || results[0].ID != "a1" {
		t.Fatalf("unexpected search results: %+v", results)
	}

	updatedAt := createdAt.Add(2 * time.Minute)
	newTitle := "Hello Go Updated"
	newTags := []string{"go", "fts"}
	updated, err := repo.UpdateArticle(ctx, "a1", domain.UpdateArticleParams{
		Title:      &newTitle,
		Tags:       &newTags,
		UpdatedAt:  updatedAt,
		IsAutoSave: true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.CurrentVersion != 2 {
		t.Fatalf("expected version 2, got %d", updated.CurrentVersion)
	}

	versions, err := repo.ListVersions(ctx, domain.ListVersionsQuery{ArticleID: "a1", Limit: 10})
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}
	if versions[0].Version != 2 || !versions[0].IsAutoSave {
		t.Fatalf("unexpected latest version: %+v", versions[0])
	}

	results, err = repo.SearchArticles(ctx, domain.SearchArticlesQuery{Query: "Updated"})
	if err != nil {
		t.Fatalf("search updated: %v", err)
	}
	if len(results) != 1 || results[0].Title != newTitle {
		t.Fatalf("unexpected search results after update: %+v", results)
	}

	restored, err := repo.RestoreVersion(ctx, "a1", 1, createdAt.Add(5*time.Minute))
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.CurrentVersion != 3 {
		t.Fatalf("expected version 3 after restore, got %d", restored.CurrentVersion)
	}
	if restored.Title != "Hello Go" {
		t.Fatalf("expected restored title, got %q", restored.Title)
	}

	allTags, err := repo.ListTags(ctx)
	if err != nil {
		t.Fatalf("list tags: %v", err)
	}
	if len(allTags) < 2 {
		t.Fatalf("expected at least 2 tags, got %d", len(allTags))
	}

	if err := repo.DeleteArticle(ctx, "a1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err = repo.GetArticle(ctx, "a1")
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestSQLiteRepository_ListArticles_ByTagAndStatus(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo, err := data.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	now := time.Now().UTC()
	_, _ = repo.CreateArticle(ctx, domain.CreateArticleParams{ID: "a1", Title: "t1", Content: "c1", Status: domain.ArticleStatusDraft, Tags: []string{"go"}, CreatedAt: now, UpdatedAt: now})
	_, _ = repo.CreateArticle(ctx, domain.CreateArticleParams{ID: "a2", Title: "t2", Content: "c2", Status: domain.ArticleStatusPublished, Tags: []string{"sql"}, CreatedAt: now, UpdatedAt: now})

	tag := "go"
	articles, err := repo.ListArticles(ctx, domain.ListArticlesQuery{Tag: &tag, Limit: 10})
	if err != nil {
		t.Fatalf("list by tag: %v", err)
	}
	if len(articles) != 1 || articles[0].ID != "a1" {
		t.Fatalf("unexpected results: %+v", articles)
	}

	status := domain.ArticleStatusPublished
	articles, err = repo.ListArticles(ctx, domain.ListArticlesQuery{Status: &status, Limit: 10})
	if err != nil {
		t.Fatalf("list by status: %v", err)
	}
	if len(articles) != 1 || articles[0].ID != "a2" {
		t.Fatalf("unexpected results: %+v", articles)
	}
}
