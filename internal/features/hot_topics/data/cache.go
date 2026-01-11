
package data

import (
	"errors"
	"sync"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type CacheEntry[V any] struct {
	Value    V
	StoredAt time.Time
	TTL      time.Duration
}

func (e CacheEntry[V]) Expired(now time.Time) bool {
	return now.After(e.StoredAt.Add(e.TTL))
}

type TtlCache[K comparable, V any] struct {
	mu    sync.Mutex
	store map[K]CacheEntry[V]
	ttl   time.Duration
	clock domain.Clock
	copy  func(V) V
}

func NewTtlCache[K comparable, V any](ttl time.Duration, clock domain.Clock, copyFn func(V) V) (*TtlCache[K, V], error) {
	if ttl <= 0 {
		return nil, errors.New("ttl cache: ttl must be positive")
	}
	if clock == nil {
		return nil, errors.New("ttl cache: clock is nil")
	}
	if copyFn == nil {
		copyFn = func(v V) V { return v }
	}
	return &TtlCache[K, V]{store: make(map[K]CacheEntry[V]), ttl: ttl, clock: clock, copy: copyFn}, nil
}

func (c *TtlCache[K, V]) Read(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.store[key]
	if !ok {
		var zero V
		return zero, false
	}
	if entry.Expired(c.clock.Now()) {
		delete(c.store, key)
		var zero V
		return zero, false
	}
	return c.copy(entry.Value), true
}

func (c *TtlCache[K, V]) Write(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = CacheEntry[V]{Value: c.copy(value), StoredAt: c.clock.Now(), TTL: c.ttl}
}

func (c *TtlCache[K, V]) Invalidate(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
}

func (c *TtlCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	clear(c.store)
}

func (c *TtlCache[K, V]) ClearExpired() {
	now := c.clock.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, entry := range c.store {
		if entry.Expired(now) {
			delete(c.store, k)
		}
	}
}

func (c *TtlCache[K, V]) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.store)
}

type HotTopicsCache struct {
	hotList *TtlCache[string, []domain.Topic]
	search  *TtlCache[string, []domain.Topic]
}

func cloneTopics(in []domain.Topic) []domain.Topic {
	if in == nil {
		return nil
	}
	out := make([]domain.Topic, len(in))
	copy(out, in)
	return out
}

func NewHotTopicsCache(ttl time.Duration, clock domain.Clock) (*HotTopicsCache, error) {
	hotList, err := NewTtlCache[string, []domain.Topic](ttl, clock, cloneTopics)
	if err != nil {
		return nil, err
	}
	search, err := NewTtlCache[string, []domain.Topic](ttl, clock, cloneTopics)
	if err != nil {
		return nil, err
	}
	return &HotTopicsCache{hotList: hotList, search: search}, nil
}

func (c *HotTopicsCache) ReadHotTopics(source *domain.Source) ([]domain.Topic, bool) {
	if c == nil {
		return nil, false
	}
	return c.hotList.Read(hotKey(source))
}

func (c *HotTopicsCache) WriteHotTopics(source *domain.Source, topics []domain.Topic) {
	if c == nil {
		return
	}
	c.hotList.Write(hotKey(source), topics)
}

func (c *HotTopicsCache) ReadSearch(source *domain.Source, query string) ([]domain.Topic, bool) {
	if c == nil {
		return nil, false
	}
	return c.search.Read(searchKey(source, query))
}

func (c *HotTopicsCache) WriteSearch(source *domain.Source, query string, topics []domain.Topic) {
	if c == nil {
		return
	}
	c.search.Write(searchKey(source, query), topics)
}

func (c *HotTopicsCache) InvalidateHotTopics(source *domain.Source) {
	if c == nil {
		return
	}
	c.hotList.Invalidate(hotKey(source))
}

func (c *HotTopicsCache) ClearSearch() {
	if c == nil {
		return
	}
	c.search.Clear()
}

func (c *HotTopicsCache) ClearAll() {
	if c == nil {
		return
	}
	c.hotList.Clear()
	c.search.Clear()
}

func hotKey(source *domain.Source) string {
	if source == nil {
		return "hot:all"
	}
	return "hot:" + source.Key()
}

func searchKey(source *domain.Source, query string) string {
	nq := domain.NormalizeQuery(query)
	if source == nil {
		return "search:all:" + nq
	}
	return "search:" + source.Key() + ":" + nq
}
