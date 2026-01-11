
package sources

import (
	"context"
	"testing"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/domain"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type fetcherFake struct {
	resp Response
	err  error
	uri  string
}

func (f *fetcherFake) Get(ctx context.Context, uri string, headers map[string]string, timeout time.Duration) (Response, error) {
	f.uri = uri
	return f.resp, f.err
}

func TestWeibo_ParseAndFetch(t *testing.T) {
	body := []byte(`{"data":{"realtime":[{"note":"Topic A","rank":1,"raw_hot":1234,"link":"https://a"},{"word":"Topic B","realpos":2,"num":200}]}}`)
	now := time.Unix(0, 0).UTC()
	topics, err := ParseWeiboHotTopics(body, now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(topics) != 2 {
		t.Fatalf("expected 2 topics, got %d", len(topics))
	}
	if topics[0].Source != domain.SourceWeibo || topics[0].Rank != 1 {
		t.Fatalf("unexpected first: %+v", topics[0])
	}

	ff := &fetcherFake{resp: Response{StatusCode: 200, Body: body}}
	src := NewWeiboSource(ff, fixedClock{t: now})
	got, err := src.FetchHotTopics(context.Background())
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 topics")
	}
	if ff.uri == "" {
		t.Fatalf("expected uri used")
	}
}

func TestZhihu_Parse(t *testing.T) {
	body := []byte(`{"data":[{"target":{"title":"Z1","excerpt":"E1","url":"https://z1"},"detail_text":"123 ??"}]}`)
	now := time.Unix(0, 0).UTC()
	topics, err := ParseZhihuHotTopics(body, now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1")
	}
	if topics[0].Source != domain.SourceZhihu || topics[0].Rank != 1 {
		t.Fatalf("unexpected: %+v", topics[0])
	}
	if topics[0].HotValue == nil || *topics[0].HotValue != 123 {
		t.Fatalf("unexpected hot value: %+v", topics[0].HotValue)
	}
}

func TestBaidu_Parse(t *testing.T) {
	body := []byte(`{"data":{"cards":[{"content":[{"word":"B1","hotScore":999,"url":"https://b1","desc":"D"}]}]}}`)
	now := time.Unix(0, 0).UTC()
	topics, err := ParseBaiduHotTopics(body, now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1")
	}
	if topics[0].Source != domain.SourceBaidu {
		t.Fatalf("unexpected source")
	}
}

func TestKr36_Parse(t *testing.T) {
	body := []byte(`{"data":{"hotRankList":[{"title":"K1","hotValue":10,"url":"https://k1"}]}}`)
	now := time.Unix(0, 0).UTC()
	topics, err := ParseKr36HotTopics(body, now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(topics) != 1 {
		t.Fatalf("expected 1")
	}
	if topics[0].Source != domain.SourceKr36 {
		t.Fatalf("unexpected source")
	}
}

func TestFetchHotTopics_StatusCodeError(t *testing.T) {
	ff := &fetcherFake{resp: Response{StatusCode: 500, Body: []byte(`{}`)}}
	clk := fixedClock{t: time.Unix(0, 0).UTC()}
	_, err := NewWeiboSource(ff, clk).FetchHotTopics(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}
