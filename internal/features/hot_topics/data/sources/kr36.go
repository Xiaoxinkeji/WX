
package sources

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type Kr36Source struct {
	Fetcher Fetcher
	HotURI  string
	Clock   domain.Clock
}

func NewKr36Source(fetcher Fetcher, clock domain.Clock) Kr36Source {
	return Kr36Source{
		Fetcher: fetcher,
		HotURI:  "https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot",
		Clock:   clock,
	}
}

func (s Kr36Source) Source() domain.Source { return domain.SourceKr36 }

func (s Kr36Source) FetchHotTopics(ctx context.Context) ([]domain.Topic, error) {
	if s.Fetcher == nil {
		return nil, errors.New("kr36 source: fetcher is nil")
	}
	if s.Clock == nil {
		return nil, errors.New("kr36 source: clock is nil")
	}
	uri := s.HotURI
	if uri == "" {
		uri = "https://gateway.36kr.com/api/mis/nav/home/nav/rank/hot"
	}
	resp, err := s.Fetcher.Get(ctx, uri, map[string]string{
		"Accept":     "application/json, text/plain, */*",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120 Safari/537.36",
	}, 15*time.Second)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("36kr hot rank failed: %d", resp.StatusCode)
	}
	return ParseKr36HotTopics(resp.Body, s.Clock.Now().UTC())
}

func ParseKr36HotTopics(body []byte, fetchedAt time.Time) ([]domain.Topic, error) {
	decoded, err := decodeJSON(body)
	if err != nil {
		return nil, err
	}
	root := asMap(decoded)

	data := asMap(root["data"])
	if data == nil {
		data = root
	}

	list := asList(data["hotRankList"])
	if list == nil {
		list = asList(data["items"])
	}
	if list == nil {
		list = asList(data["list"])
	}
	if list == nil {
		list = asList(data["data"])
	}

	if list == nil {
		inner := asMap(data["data"])
		list = asList(inner["hotRankList"])
		if list == nil {
			list = asList(inner["items"])
		}
		if list == nil {
			list = asList(inner["list"])
		}
	}
	if list == nil {
		list = []any{}
	}

	out := make([]domain.Topic, 0, len(list))
	for i, raw := range list {
		item := asMap(raw)
		if item == nil {
			continue
		}

		title, _ := asString(item["title"])
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["name"])
		}
		if strings.TrimSpace(title) == "" {
			title, _ = asString(item["word"])
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
		if v, ok := asFloat(item["hotValue"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["score"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["hot"]); ok {
			hotValue = &v
		} else if v, ok := asFloat(item["hotRank"]); ok {
			hotValue = &v
		} else if t, ok := asString(item["hotValue"]); ok {
			if v, ok := tryParseFloatFromText(t); ok {
				hotValue = &v
			}
		}

		desc, _ := asString(item["desc"])
		if strings.TrimSpace(desc) == "" {
			desc, _ = asString(item["summary"])
		}
		if strings.TrimSpace(desc) == "" {
			desc, _ = asString(item["description"])
		}

		topic, err := domain.NewTopicFromParts(domain.SourceKr36, i+1, title, maybeStringPtr(urlStr), hotValue, maybeStringPtr(desc), fetchedAt)
		if err != nil {
			continue
		}
		out = append(out, topic)
	}

	return out, nil
}
