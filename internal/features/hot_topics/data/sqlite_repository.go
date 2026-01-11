
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/data/models"
	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/data/sources"
	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now().UTC() }

type SQLiteRepository struct {
	db      *sql.DB
	sources []sources.HotTopicSource
	cache   *HotTopicsCache
	clock   domain.Clock
	ttl     time.Duration
}

type Option func(*SQLiteRepository) error

func WithSources(srcs ...sources.HotTopicSource) Option {
	return func(r *SQLiteRepository) error {
		if len(srcs) == 0 {
			r.sources = nil
			return nil
		}
		out := make([]sources.HotTopicSource, 0, len(srcs))
		for _, s := range srcs {
			if s != nil {
				out = append(out, s)
			}
		}
		r.sources = out
		return nil
	}
}

func WithClock(clock domain.Clock) Option {
	return func(r *SQLiteRepository) error {
		if clock == nil {
			return errors.New("sqlite repository: clock is nil")
		}
		r.clock = clock
		return nil
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(r *SQLiteRepository) error {
		if ttl <= 0 {
			return errors.New("sqlite repository: ttl must be positive")
		}
		r.ttl = ttl
		return nil
	}
}

func WithCache(cache *HotTopicsCache) Option {
	return func(r *SQLiteRepository) error {
		r.cache = cache
		return nil
	}
}

func NewSQLiteRepository(db *sql.DB, opts ...Option) (*SQLiteRepository, error) {
	if db == nil {
		return nil, errors.New("sqlite repository: db is nil")
	}

	repo := &SQLiteRepository{db: db, ttl: 10 * time.Minute, clock: systemClock{}}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(repo); err != nil {
			return nil, err
		}
	}
	if repo.clock == nil {
		repo.clock = systemClock{}
	}
	if repo.ttl <= 0 {
		repo.ttl = 10 * time.Minute
	}
	if repo.cache == nil {
		c, err := NewHotTopicsCache(repo.ttl, repo.clock)
		if err != nil {
			return nil, err
		}
		repo.cache = c
	}
	if len(repo.sources) == 0 {
		fetcher := DefaultAPIClient{UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36"}
		repo.sources = []sources.HotTopicSource{
			sources.NewWeiboSource(fetcher, repo.clock),
			sources.NewZhihuSource(fetcher, repo.clock),
			sources.NewBaiduSource(fetcher, repo.clock),
			sources.NewKr36Source(fetcher, repo.clock),
		}
	}

	if err := repo.EnsureSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteRepository) EnsureSchema(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS hot_topics (
	id TEXT PRIMARY KEY,
	source TEXT NOT NULL,
	rank INTEGER NOT NULL,
	title TEXT NOT NULL,
	url TEXT,
	hot_value REAL,
	description TEXT,
	fetched_at_ms INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_hot_topics_source_rank ON hot_topics(source, rank);
CREATE INDEX IF NOT EXISTS idx_hot_topics_source_fetched ON hot_topics(source, fetched_at_ms DESC);
`)
	return err
}

func (r *SQLiteRepository) GetHotTopics(ctx context.Context, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("get hot topics: repo is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if source != nil && !source.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid source"))
	}

	if !forceRefresh {
		if cached, ok := r.cache.ReadHotTopics(source); ok {
			return cached, nil
		}
		if topics, ok, err := r.readFreshFromSQLite(ctx, source); err != nil {
			return nil, err
		} else if ok {
			r.cache.WriteHotTopics(source, topics)
			return topics, nil
		}
	}

	if source != nil {
		topics, err := r.fetchOne(ctx, *source)
		if err != nil {
			if !forceRefresh {
				if fallback, ok, err2 := r.readAnyFromSQLite(ctx, source); err2 == nil && ok {
					r.cache.WriteHotTopics(source, fallback)
					return fallback, nil
				}
			}
			return nil, errors.Join(domain.ErrProvider, err)
		}
		if err := r.replaceTopics(ctx, *source, topics); err != nil {
			return nil, err
		}
		r.cache.WriteHotTopics(source, topics)
		return topics, nil
	}

	attempts := r.fetchAll(ctx)
	merged := mergeAcrossSources(r.sources, attempts)

	if len(merged) == 0 {
		var firstErr error
		for _, a := range attempts {
			if a.err != nil {
				firstErr = a.err
				break
			}
		}
		if firstErr != nil {
			if !forceRefresh {
				if fallback, ok, err2 := r.readAnyFromSQLite(ctx, nil); err2 == nil && ok {
					r.cache.WriteHotTopics(nil, fallback)
					return fallback, nil
				}
			}
			return nil, errors.Join(domain.ErrProvider, firstErr)
		}
	}

	if err := r.replaceMany(ctx, attempts); err != nil {
		return nil, err
	}

	r.cache.WriteHotTopics(nil, merged)
	for _, a := range attempts {
		if a.err == nil && len(a.topics) > 0 {
			s := a.source
			r.cache.WriteHotTopics(&s, a.topics)
		}
	}
	return merged, nil
}

func (r *SQLiteRepository) RefreshHotTopics(ctx context.Context, source *domain.Source) ([]domain.Topic, error) {
	if r.cache != nil {
		r.cache.InvalidateHotTopics(source)
		r.cache.ClearSearch()
	}
	return r.GetHotTopics(ctx, source, true)
}

func (r *SQLiteRepository) SearchHotTopics(ctx context.Context, query string, source *domain.Source, forceRefresh bool) ([]domain.Topic, error) {
	if r == nil {
		return nil, errors.New("search hot topics: repo is nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if source != nil && !source.Valid() {
		return nil, errors.Join(domain.ErrInvalidArgument, errors.New("invalid source"))
	}

	q := strings.TrimSpace(query)
	if q == "" {
		return r.GetHotTopics(ctx, source, forceRefresh)
	}

	if !forceRefresh {
		if cached, ok := r.cache.ReadSearch(source, q); ok {
			return cached, nil
		}
	}

	base, err := r.GetHotTopics(ctx, source, forceRefresh)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Topic, 0)
	for _, t := range base {
		if domain.ContainsIgnoreCase(t.Title, q) {
			out = append(out, t)
			continue
		}
		if t.Description != nil && domain.ContainsIgnoreCase(*t.Description, q) {
			out = append(out, t)
		}
	}

	r.cache.WriteSearch(source, q, out)
	return out, nil
}

type fetchAttempt struct {
	source domain.Source
	topics []domain.Topic
	err    error
}

func (r *SQLiteRepository) fetchOne(ctx context.Context, source domain.Source) ([]domain.Topic, error) {
	adapter, err := r.findSource(source)
	if err != nil {
		return nil, err
	}
	topics, err := adapter.FetchHotTopics(ctx)
	if err != nil {
		return nil, err
	}
	return sortedForSingleSource(topics), nil
}

func (r *SQLiteRepository) fetchAll(ctx context.Context) []fetchAttempt {
	attempts := make([]fetchAttempt, len(r.sources))
	var wg sync.WaitGroup
	wg.Add(len(r.sources))
	for i, src := range r.sources {
		i := i
		src := src
		go func() {
			defer wg.Done()
			if src == nil {
				attempts[i] = fetchAttempt{err: errors.New("source is nil")}
				return
			}
			topics, err := src.FetchHotTopics(ctx)
			if err != nil {
				attempts[i] = fetchAttempt{source: src.Source(), topics: nil, err: err}
				return
			}
			attempts[i] = fetchAttempt{source: src.Source(), topics: sortedForSingleSource(topics), err: nil}
		}()
	}
	wg.Wait()
	return attempts
}

func mergeAcrossSources(order []sources.HotTopicSource, attempts []fetchAttempt) []domain.Topic {
	sourceOrder := make(map[domain.Source]int, len(order))
	for i, s := range order {
		if s == nil {
			continue
		}
		sourceOrder[s.Source()] = i
	}

	merged := make([]domain.Topic, 0)
	for _, a := range attempts {
		merged = append(merged, a.topics...)
	}

	sort.Slice(merged, func(i, j int) bool {
		ao, ok := sourceOrder[merged[i].Source]
		if !ok {
			ao = 999
		}
		bo, ok := sourceOrder[merged[j].Source]
		if !ok {
			bo = 999
		}
		if ao != bo {
			return ao < bo
		}
		return merged[i].Rank < merged[j].Rank
	})
	return merged
}

func sortedForSingleSource(topics []domain.Topic) []domain.Topic {
	copyTopics := make([]domain.Topic, len(topics))
	copy(copyTopics, topics)
	sort.Slice(copyTopics, func(i, j int) bool { return copyTopics[i].Rank < copyTopics[j].Rank })
	return copyTopics
}

func (r *SQLiteRepository) findSource(source domain.Source) (sources.HotTopicSource, error) {
	for _, s := range r.sources {
		if s != nil && s.Source() == source {
			return s, nil
		}
	}
	return nil, errors.Join(domain.ErrInvalidArgument, fmt.Errorf("no source registered for %s", source.Key()))
}

func (r *SQLiteRepository) replaceTopics(ctx context.Context, source domain.Source, topics []domain.Topic) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := replaceTopicsTx(ctx, tx, source, topics); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLiteRepository) replaceMany(ctx context.Context, attempts []fetchAttempt) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, a := range attempts {
		if a.err != nil {
			continue
		}
		if err := replaceTopicsTx(ctx, tx, a.source, a.topics); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func replaceTopicsTx(ctx context.Context, tx *sql.Tx, source domain.Source, topics []domain.Topic) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM hot_topics WHERE source = ?`, source.Key()); err != nil {
		return err
	}
	if len(topics) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO hot_topics(id, source, rank, title, url, hot_value, description, fetched_at_ms)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, t := range topics {
		if !t.Source.Valid() || t.Source != source {
			continue
		}
		dto := models.TopicDTOFromDomain(t)
		if _, err := stmt.ExecContext(ctx, dto.ID, dto.Source, dto.Rank, dto.Title, nullStringArg(dto.URL), nullFloatArg(dto.HotValue), nullStringArg(dto.Description), dto.FetchedAtMs); err != nil {
			return err
		}
	}
	return nil
}

func nullStringArg(v sql.NullString) any {
	if v.Valid {
		return v.String
	}
	return nil
}

func nullFloatArg(v sql.NullFloat64) any {
	if v.Valid {
		return v.Float64
	}
	return nil
}

func (r *SQLiteRepository) readFreshFromSQLite(ctx context.Context, source *domain.Source) ([]domain.Topic, bool, error) {
	now := r.clock.Now().UTC()
	thresholdMs := now.Add(-r.ttl).UnixMilli()

	if source != nil {
		var max sql.NullInt64
		if err := r.db.QueryRowContext(ctx, `SELECT MAX(fetched_at_ms) FROM hot_topics WHERE source = ?`, source.Key()).Scan(&max); err != nil {
			return nil, false, err
		}
		if !max.Valid {
			return nil, false, nil
		}
		if max.Int64 < thresholdMs {
			return nil, false, nil
		}
		topics, err := r.listFromSQLite(ctx, source)
		if err != nil {
			return nil, false, err
		}
		return topics, len(topics) > 0, nil
	}

	rows, err := r.db.QueryContext(ctx, `SELECT source, MAX(fetched_at_ms) FROM hot_topics GROUP BY source`)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	maxBySource := make(map[string]int64)
	for rows.Next() {
		var s string
		var ms sql.NullInt64
		if err := rows.Scan(&s, &ms); err != nil {
			return nil, false, err
		}
		if ms.Valid {
			maxBySource[s] = ms.Int64
		}
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	for _, src := range r.expectedSources() {
		ms, ok := maxBySource[src.Key()]
		if !ok || ms < thresholdMs {
			return nil, false, nil
		}
	}

	topics, err := r.listFromSQLite(ctx, nil)
	if err != nil {
		return nil, false, err
	}
	return topics, len(topics) > 0, nil
}

func (r *SQLiteRepository) readAnyFromSQLite(ctx context.Context, source *domain.Source) ([]domain.Topic, bool, error) {
	topics, err := r.listFromSQLite(ctx, source)
	if err != nil {
		return nil, false, err
	}
	return topics, len(topics) > 0, nil
}

func (r *SQLiteRepository) listFromSQLite(ctx context.Context, source *domain.Source) ([]domain.Topic, error) {
	q := `
SELECT id, source, rank, title, url, hot_value, description, fetched_at_ms
FROM hot_topics
`
	args := []any{}
	if source != nil {
		q += " WHERE source = ?
"
		args = append(args, source.Key())
		q += " ORDER BY rank ASC
"
	} else {
		q += " ORDER BY " + sourceCaseOrder("source", r.expectedSources()) + ", rank ASC
"
	}

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Topic, 0)
	for rows.Next() {
		var dto models.TopicDTO
		if err := rows.Scan(&dto.ID, &dto.Source, &dto.Rank, &dto.Title, &dto.URL, &dto.HotValue, &dto.Description, &dto.FetchedAtMs); err != nil {
			return nil, err
		}
		t, err := dto.ToDomain()
		if err != nil {
			continue
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SQLiteRepository) expectedSources() []domain.Source {
	seen := make(map[domain.Source]struct{})
	out := make([]domain.Source, 0, len(r.sources))
	for _, s := range r.sources {
		if s == nil {
			continue
		}
		src := s.Source()
		if !src.Valid() {
			continue
		}
		if _, ok := seen[src]; ok {
			continue
		}
		seen[src] = struct{}{}
		out = append(out, src)
	}
	if len(out) == 0 {
		return domain.AllSources()
	}
	return out
}

func sourceCaseOrder(column string, sources []domain.Source) string {
	var b strings.Builder
	b.WriteString("CASE ")
	b.WriteString(column)
	for i, s := range sources {
		b.WriteString(" WHEN '")
		b.WriteString(s.Key())
		b.WriteString("' THEN ")
		b.WriteString(fmt.Sprintf("%d", i))
	}
	b.WriteString(" ELSE ")
	b.WriteString(fmt.Sprintf("%d", len(sources)))
	b.WriteString(" END")
	return b.String()
}
