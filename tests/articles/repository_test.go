package articles_test

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

func openRepo(t *testing.T) (context.Context, *data.SQLiteRepository) {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("file:articles_repo_%d?mode=memory&cache=shared", time.Now().UnixNano())
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
	return ctx, repo
}

func TestRepository_NotFound(t *testing.T) {
	ctx, repo := openRepo(t)
	_, err := repo.GetArticle(ctx, "missing")
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if err := repo.DeleteArticle(ctx, "missing"); err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRepository_InvalidArguments(t *testing.T) {
	ctx, repo := openRepo(t)
	_, err := repo.CreateArticle(ctx, domain.CreateArticleParams{ID: ""})
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	_, err = repo.UpdateArticle(ctx, "", domain.UpdateArticleParams{})
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
