
package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type BaiduSource struct {
	Fetcher Fetcher
	HotURI  string
	Clock   domain.Clock
}

func NewBaiduSource(fetcher Fetcher, clock domain.Clock) BaiduSource {
	return BaiduSource{
		Fetcher: fetcher,
		HotURI:  "https://top.baidu.com/api/board?platform=wise&tab=realtime",
		Clock:   clock,
	}
}

func (s BaiduSource) Source() domain.Source { return domain.SourceBaidu }

func (s BaiduSource) FetchHotTopics(ctx context.Context) ([]domain.Topic, error) {
	if s.Fetcher == nil {
		return nil, errors.New("baidu source: fetcher is nil")
	}
	if s.Clock == nil {
		return nil, errors.New("baidu source: clock is nil")
	}
	uri := s.HotURI
	if uri == "" {
		uri = "https://top.baidu.com/api/board?platform=wise&tab=realtime"
	}
	resp, err := s.Fetcher.Get(ctx, uri, map[string]string{
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36",
	}, 15*time.Second)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("baidu hot board failed: %d", resp.StatusCode)
	}
	return ParseBaiduHotTopics(resp.Body, s.Clock.Now().UTC())
}

func ParseBaiduHotTopics(body []byte, fetchedAt time.Time) ([]domain.Topic, error) {
	decoded, err := decodeJSON(body)
	if err != nil {
		return nil, err
	}
	root := asMap(decoded)
	data := asMap(root["data"])
	cards := asList(data["cards"])
	if cards == nil {
		cards = []any{}
	}

	entries := make([]map[string]any, 0)
	for _, rawCard := range cards {
		card := asMap(rawCard)
		if card == nil {
			continue
		}
		content := asList(card["content"])
		for _, rawItem := range content {
			m := asMap(rawItem)
			if m != nil {
				entries = append(entries, m)
			}
		}
	}

	out := make([]domain.Topic, 0, len(entries))
	for i, item := range entries {
		title, _ := asString(item["word"])
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["keyword"])
		}
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["title"])
		}
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		var urlStr string
		if u, ok := asString(item["url"]); ok {
			urlStr = u
		} else if u, ok := asString(item["link"]); ok {
			urlStr = u
		}

		var hotValue *float64
		if v, ok := asFloat(item["hotScore"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["hot_score"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["hotValue"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["score"]); ok {
			hotValue = &v
		} else if t, ok := asString(item["hotScore"]); ok {
			if v, ok := tryParseFloatFromText(t); ok {
				hotValue = &v
			}
		}

		desc, _ := asString(item["desc"])
		if strings.TrimSpace(desc) == "" {
			desc, _ = asString(item["desc1"])
		}
		if strings.TrimSpace(desc) == "" {
			desc, _ = asString(item["summary"])
		}

		topic, err := domain.NewTopicFromParts(domain.SourceBaidu, i+1, title, maybeStringPtr(urlStr), hotValue, maybeStringPtr(desc), fetchedAt)
		if err != nil {
			continue
		}
		out = append(out, topic)
	}

	return out, nil
}
