
package sources

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type WeiboSource struct {
	Fetcher Fetcher
	HotURI  string
	Clock   domain.Clock
}

func NewWeiboSource(fetcher Fetcher, clock domain.Clock) WeiboSource {
	return WeiboSource{
		Fetcher: fetcher,
		HotURI:  "https://weibo.com/ajax/side/hotSearch",
		Clock:   clock,
	}
}

func (s WeiboSource) Source() domain.Source { return domain.SourceWeibo }

func (s WeiboSource) FetchHotTopics(ctx context.Context) ([]domain.Topic, error) {
	if s.Fetcher == nil {
		return nil, errors.New("weibo source: fetcher is nil")
	}
	if s.Clock == nil {
		return nil, errors.New("weibo source: clock is nil")
	}
	uri := s.HotURI
	if uri == "" {
		uri = "https://weibo.com/ajax/side/hotSearch"
	}
	resp, err := s.Fetcher.Get(ctx, uri, map[string]string{
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36",
	}, 15*time.Second)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("weibo hotSearch failed: %d", resp.StatusCode)
	}
	return ParseWeiboHotTopics(resp.Body, s.Clock.Now().UTC())
}

func ParseWeiboHotTopics(body []byte, fetchedAt time.Time) ([]domain.Topic, error) {
	decoded, err := decodeJSON(body)
	if err != nil {
		return nil, err
	}
	root := asMap(decoded)
	data := asMap(root["data"])
	realtime := asList(data["realtime"])
	if realtime == nil {
		realtime = []any{}
	}

	out := make([]domain.Topic, 0, len(realtime))
	for i, raw := range realtime {
		item := asMap(raw)
		if item == nil {
			continue
		}

		title, _ := asString(item["note"])
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["word"])
		}
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		rank, ok := asInt(item["rank"])
		if !ok {
			rank, ok = asInt(item["realpos"])
		}
		if !ok {
			rank, ok = asInt(item["num"])
		}
		if !ok {
			rank = i + 1
		}

		var urlStr string
		if u, ok := asString(item["link"]); ok {
			urlStr = u
		} else if u, ok := asString(item["url"]); ok {
			urlStr = u
		}

		var hotValue *float64
		if v, ok := asFloat(item["raw_hot"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["rawHot"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["hot"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["num"]); ok {
			hotValue = &v
		} else if raw, ok := asString(item["raw_hot"]); ok {
			if v, ok := tryParseFloatFromText(raw); ok {
				hotValue = &v
			}
		}

		topic, err := domain.NewTopicFromParts(domain.SourceWeibo, rank, title, maybeStringPtr(urlStr), hotValue, nil, fetchedAt)
		if err != nil {
			continue
		}
		out = append(out, topic)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Rank < out[j].Rank })
	return out, nil
}
