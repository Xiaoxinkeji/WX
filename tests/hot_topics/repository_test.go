
package hottopics_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/data"
	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type sourceFake struct {
	src    domain.Source
	topics []domain.Topic
	err    error
	calls  int
}

func (s *sourceFake) Source() domain.Source { return s.src }

func (s *sourceFake) FetchHotTopics(ctx context.Context) ([]domain.Topic, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.topics, nil
}

func openRepo(t *testing.T) (context.Context, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	name := fmt.Sprintf("file:hot_topics_repo_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return ctx, db
}

func TestRepository_InvalidSource(t *testing.T) {
	ctx, db := openRepo(t)
	clk := fixedClock{t: time.Unix(0, 0).UTC()}
	repo, err := data.NewSQLiteRepository(db, data.WithClock(clk), data.WithSources(&sourceFake{src: domain.SourceWeibo}))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	bad := domain.Source("bad")
	_, err = repo.GetHotTopics(ctx, &bad, false)
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestRepository_RefreshAndSearch(t *testing.T) {
	ctx, db := openRepo(t)
	clk := fixedClock{t: time.Unix(0, 0).UTC()}

	src := domain.SourceWeibo
	topics := []domain.Topic{{ID: "weibo:x", Source: src, Rank: 1, Title: "Hello", FetchedAt: clk.Now()}}
	remote := &sourceFake{src: src, topics: topics}

	repo, err := data.NewSQLiteRepository(db, data.WithClock(clk), data.WithTTL(1*time.Hour), data.WithSources(remote))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	_, err = repo.RefreshHotTopics(ctx, &src)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}

	got, err := repo.SearchHotTopics(ctx, "hell", &src, false)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1")
	}
}
