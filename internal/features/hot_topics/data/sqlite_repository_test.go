
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
	srcpkg "github.com/Xiaoxinkeji/WX/internal/features/hot_topics/data/sources"
)

type movingClock struct{ t time.Time }

func (c *movingClock) Now() time.Time { return c.t }

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

func openHotTopicsDB(t *testing.T) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("file:hot_topics_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestSQLiteRepository_GetHotTopics_CacheThenSQLiteThenFallback(t *testing.T) {
	ctx := context.Background()
	db := openHotTopicsDB(t)
	clk := &movingClock{t: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}

	src := domain.SourceWeibo
	seed := []domain.Topic{{ID: "weibo:x", Source: src, Rank: 2, Title: "B", FetchedAt: clk.Now()}}
	seed = append(seed, domain.Topic{ID: "weibo:y", Source: src, Rank: 1, Title: "A", FetchedAt: clk.Now()})
	remote := &sourceFake{src: src, topics: seed}

	repo, err := NewSQLiteRepository(db, WithClock(clk), WithTTL(1*time.Hour), WithSources(remote))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	got, err := repo.GetHotTopics(ctx, &src, false)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if remote.calls != 1 {
		t.Fatalf("expected 1 remote call, got %d", remote.calls)
	}
	if len(got) != 2 || got[0].Rank != 1 {
		t.Fatalf("expected sorted topics, got %+v", got)
	}

	_, _ = repo.GetHotTopics(ctx, &src, false)
	if remote.calls != 1 {
		t.Fatalf("expected cache hit")
	}

	remote2 := &sourceFake{src: src, err: errors.New("network down")}
	repo2, err := NewSQLiteRepository(db, WithClock(clk), WithTTL(1*time.Hour), WithSources(remote2))
	if err != nil {
		t.Fatalf("new repo2: %v", err)
	}
	got2, err := repo2.GetHotTopics(ctx, &src, false)
	if err != nil {
		t.Fatalf("get from sqlite: %v", err)
	}
	if remote2.calls != 0 {
		t.Fatalf("expected no remote call when sqlite fresh")
	}
	if len(got2) != 2 {
		t.Fatalf("unexpected sqlite topics")
	}

	clk.t = clk.t.Add(2 * time.Hour)
	repo3, err := NewSQLiteRepository(db, WithClock(clk), WithTTL(30*time.Minute), WithSources(remote2))
	if err != nil {
		t.Fatalf("new repo3: %v", err)
	}
	got3, err := repo3.GetHotTopics(ctx, &src, false)
	if err != nil {
		t.Fatalf("expected stale sqlite fallback, got %v", err)
	}
	if len(got3) != 2 {
		t.Fatalf("expected fallback topics")
	}
}

func TestSQLiteRepository_GetHotTopics_AllSources_MergeAndPartialFailure(t *testing.T) {
	ctx := context.Background()
	db := openHotTopicsDB(t)
	clk := &movingClock{t: time.Unix(0, 0).UTC()}

	w := domain.SourceWeibo
	z := domain.SourceZhihu

	weibo := &sourceFake{src: w, topics: []domain.Topic{{ID: "weibo:a", Source: w, Rank: 2, Title: "w2", FetchedAt: clk.Now()}, {ID: "weibo:b", Source: w, Rank: 1, Title: "w1", FetchedAt: clk.Now()}}}
	zhihu := &sourceFake{src: z, err: errors.New("boom")}

	repo, err := NewSQLiteRepository(db, WithClock(clk), WithTTL(1*time.Hour), WithSources(weibo, zhihu))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	got, err := repo.GetHotTopics(ctx, nil, true)
	if err != nil {
		t.Fatalf("expected partial success, got %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected topics from weibo only")
	}
	if got[0].Source != w || got[0].Rank != 1 {
		t.Fatalf("unexpected merge order: %+v", got)
	}
}

func TestSQLiteRepository_SearchHotTopics_UsesCache(t *testing.T) {
	ctx := context.Background()
	db := openHotTopicsDB(t)
	clk := &movingClock{t: time.Unix(0, 0).UTC()}

	src := domain.SourceWeibo
	remote := &sourceFake{src: src, topics: []domain.Topic{{ID: "weibo:x", Source: src, Rank: 1, Title: "Hello", FetchedAt: clk.Now()}}}
	repo, err := NewSQLiteRepository(db, WithClock(clk), WithTTL(1*time.Hour), WithSources(remote))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	got, err := repo.SearchHotTopics(ctx, "hello", &src, false)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1")
	}

	remote.calls = 0
	got2, err := repo.SearchHotTopics(ctx, "hello", &src, false)
	if err != nil {
		t.Fatalf("search2: %v", err)
	}
	if len(got2) != 1 {
		t.Fatalf("expected 1")
	}
	if remote.calls != 0 {
		t.Fatalf("expected search cache hit")
	}
}

func TestSQLiteRepository_InvalidSource(t *testing.T) {
	db := openHotTopicsDB(t)
	repo, err := NewSQLiteRepository(db, WithSources(&sourceFake{src: domain.SourceWeibo}))
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}
	bad := domain.Source("bad")
	_, err = repo.GetHotTopics(context.Background(), &bad, false)
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestDefaultAPIClient_ImplementsFetcher(t *testing.T) {
	var _ srcpkg.Fetcher = DefaultAPIClient{}
}
