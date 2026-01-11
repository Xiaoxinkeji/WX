
package data

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/hot_topics/data/sources"
)

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultAPIClient struct {
	HTTPClient   HTTPDoer
	UserAgent    string
	MaxBodyBytes int64
}

func (c DefaultAPIClient) Get(ctx context.Context, uri string, headers map[string]string, timeout time.Duration) (sources.Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return sources.Response{}, errors.New("api client: uri is empty")
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return sources.Response{}, err
	}
	for k, v := range headers {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	if ua := strings.TrimSpace(c.UserAgent); ua != "" && req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", ua)
	}

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	resp, err := client.Do(req)
	if err != nil {
		return sources.Response{}, err
	}
	defer resp.Body.Close()

	limit := c.MaxBodyBytes
	if limit <= 0 {
		limit = 4 * 1024 * 1024
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	if err != nil {
		return sources.Response{}, err
	}

	h := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		if len(v) > 0 {
			h[k] = v[0]
		}
	}
	return sources.Response{StatusCode: resp.StatusCode, Body: b, Headers: h}, nil
}
