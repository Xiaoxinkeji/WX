package domain

import (
"context"
"errors"
"time"
)

var (
ErrInvalidArgument = errors.New("hot_topics: invalid argument")
ErrProvider        = errors.New("hot_topics: provider error")
ErrNotFound        = errors.New("hot_topics: not found")
)

type Clock interface {
Now() time.Time
}

type Repository interface {
GetHotTopics(ctx context.Context, source *Source, forceRefresh bool) ([]Topic, error)
RefreshHotTopics(ctx context.Context, source *Source) ([]Topic, error)
SearchHotTopics(ctx context.Context, query string, source *Source, forceRefresh bool) ([]Topic, error)
}
