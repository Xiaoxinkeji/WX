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
	"github.com/Xiaoxinkeji/WX/internal/features/articles/usecase"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type fixedIDs struct{ id string }

func (g fixedIDs) NewID() (string, error) { return g.id, nil }

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("file:articles_suite_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestUseCases_EndToEnd(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo, err := data.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	now := time.Date(2025, 3, 4, 5, 6, 7, 0, time.UTC)
	createUC := usecase.CreateArticleUseCase{Repo: repo, Clock: fixedClock{t: now}, IDs: fixedIDs{id: "id-1"}}
	updateUC := usecase.UpdateArticleUseCase{Repo: repo, Clock: fixedClock{t: now.Add(time.Minute)}}
	listUC := usecase.NewListArticlesUseCase(repo)
	searchUC := usecase.NewSearchArticlesUseCase(repo)
	deleteUC := usecase.NewDeleteArticleUseCase(repo)

	created, err := createUC.Execute(ctx, usecase.CreateArticleInput{Title: "Draft", Content: "Hello world", Tags: []string{"Go"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != "id-1" {
		t.Fatalf("expected id-1, got %q", created.ID)
	}
	if created.Status != domain.ArticleStatusDraft {
		t.Fatalf("expected draft, got %q", created.Status)
	}

	newContent := "Hello updated world"
	_, err = updateUC.Execute(ctx, usecase.UpdateArticleInput{ID: created.ID, Content: &newContent, AutoSave: true})
	if err != nil {
		t.Fatalf("update autosave: %v", err)
	}

	articles, err := listUC.Execute(ctx, usecase.ListArticlesInput{Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(articles))
	}

	found, err := searchUC.Execute(ctx, usecase.SearchArticlesInput{Query: "updated"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(found) != 1 || found[0].ID != created.ID {
		t.Fatalf("unexpected search results: %+v", found)
	}

	versions, err := repo.ListVersions(ctx, domain.ListVersionsQuery{ArticleID: created.ID, Limit: 10})
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(versions))
	}

	if err := deleteUC.Execute(ctx, usecase.DeleteArticleInput{ID: created.ID}); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
