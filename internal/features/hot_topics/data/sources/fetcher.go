
package sources

import (
	"context"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string]string
}

type Fetcher interface {
	Get(ctx context.Context, uri string, headers map[string]string, timeout time.Duration) (Response, error)
}

type HotTopicSource interface {
	Source() domain.Source
	FetchHotTopics(ctx context.Context) ([]domain.Topic, error)
}
