
package data

import (
	"testing"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time { return c.t }

func TestTtlCache_ReadWriteExpire(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0).UTC()}
	cache, err := NewTtlCache[string, int](2*time.Second, clk, nil)
	if err != nil {
		t.Fatalf("new ttl cache: %v", err)
	}

	cache.Write("k", 1)
	v, ok := cache.Read("k")
	if !ok || v != 1 {
		t.Fatalf("expected 1, got %v (ok=%v)", v, ok)
	}

	clk.t = clk.t.Add(3 * time.Second)
	_, ok = cache.Read("k")
	if ok {
		t.Fatalf("expected expired")
	}
	if cache.Size() != 0 {
		t.Fatalf("expected entry removed")
	}
}

func TestTtlCache_ClearExpired(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0).UTC()}
	cache, _ := NewTtlCache[string, int](1*time.Second, clk, nil)
	cache.Write("a", 1)
	cache.Write("b", 2)

	clk.t = clk.t.Add(2 * time.Second)
	cache.ClearExpired()
	if cache.Size() != 0 {
		t.Fatalf("expected all expired")
	}
}

func TestHotTopicsCache_KeysAndClone(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0).UTC()}
	c, err := NewHotTopicsCache(1*time.Hour, clk)
	if err != nil {
		t.Fatalf("new cache: %v", err)
	}

	src := domain.SourceWeibo
	topics := []domain.Topic{{ID: "x", Source: src, Rank: 1, Title: "T", FetchedAt: clk.Now()}}
	c.WriteHotTopics(&src, topics)

	got, ok := c.ReadHotTopics(&src)
	if !ok || len(got) != 1 {
		t.Fatalf("unexpected read: %v %v", got, ok)
	}
	got[0].Title = "mutated"

	got2, ok := c.ReadHotTopics(&src)
	if !ok || got2[0].Title != "T" {
		t.Fatalf("expected clone protection, got %q", got2[0].Title)
	}

	c.WriteSearch(nil, "  HeLLo ", topics)
	got3, ok := c.ReadSearch(nil, "hello")
	if !ok || len(got3) != 1 {
		t.Fatalf("unexpected search cache")
	}

	c.InvalidateHotTopics(&src)
	_, ok = c.ReadHotTopics(&src)
	if ok {
		t.Fatalf("expected invalidated")
	}

	c.ClearAll()
	_, ok = c.ReadSearch(nil, "hello")
	if ok {
		t.Fatalf("expected cleared")
	}
}
