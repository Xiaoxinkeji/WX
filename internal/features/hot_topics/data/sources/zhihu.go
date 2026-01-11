
package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type ZhihuSource struct {
	Fetcher Fetcher
	HotURI  string
	Clock   domain.Clock
}

func NewZhihuSource(fetcher Fetcher, clock domain.Clock) ZhihuSource {
	return ZhihuSource{
		Fetcher: fetcher,
		HotURI:  "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true",
		Clock:   clock,
	}
}

func (s ZhihuSource) Source() domain.Source { return domain.SourceZhihu }

func (s ZhihuSource) FetchHotTopics(ctx context.Context) ([]domain.Topic, error) {
	if s.Fetcher == nil {
		return nil, errors.New("zhihu source: fetcher is nil")
	}
	if s.Clock == nil {
		return nil, errors.New("zhihu source: clock is nil")
	}
	uri := s.HotURI
	if uri == "" {
		uri = "https://www.zhihu.com/api/v3/feed/topstory/hot-lists/total?limit=50&desktop=true"
	}
	resp, err := s.Fetcher.Get(ctx, uri, map[string]string{
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36",
	}, 15*time.Second)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("zhihu hot list failed: %d", resp.StatusCode)
	}
	return ParseZhihuHotTopics(resp.Body, s.Clock.Now().UTC())
}

func ParseZhihuHotTopics(body []byte, fetchedAt time.Time) ([]domain.Topic, error) {
	decoded, err := decodeJSON(body)
	if err != nil {
		return nil, err
	}
	root := asMap(decoded)
	items := asList(root["data"])
	if items == nil {
		items = []any{}
	}

	out := make([]domain.Topic, 0, len(items))
	for i, raw := range items {
		item := asMap(raw)
		if item == nil {
			continue
		}
		target := asMap(item["target"])

		title, _ := asString(target["title"])
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["title"])
		}
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		excerpt, _ := asString(target["excerpt"])
		if strings.TrimSpace(excerpt) == "" {
			excerpt, _ = asString(target["excerpt_new"])
		}
		if strings.TrimSpace(excerpt) == "" {
			excerpt, _ = asString(target["description"])
		}

		var urlStr string
		if u, ok := asString(target["url"]); ok {
			urlStr = u
		} else if u, ok := asString(item["url"]); ok {
			urlStr = u
		} else if u, ok := asString(target["url_token"]); ok {
			urlStr = u
		} else if u, ok := asString(target["urlToken"]); ok {
			urlStr = u
		}

		var hotValue *float64
		if t, ok := asString(item["detail_text"]); ok {
			if v, ok := tryParseFloatFromText(t); ok {
				hotValue = &v
			}
		} else if t, ok := asString(item["detailText"]); ok {
			if v, ok := tryParseFloatFromText(t); ok {
				hotValue = &v
			}
		} else if t, ok := asString(item["heat"]); ok {
			if v, ok := tryParseFloatFromText(t); ok {
				hotValue = &v
			}
		}

		topic, err := domain.NewTopicFromParts(domain.SourceZhihu, i+1, title, maybeStringPtr(urlStr), hotValue, maybeStringPtr(excerpt), fetchedAt)
		if err != nil {
			continue
		}
		out = append(out, topic)
	}
	return out, nil
}
