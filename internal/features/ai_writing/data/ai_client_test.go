package data_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

func TestOpenAIClient_Chat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		b, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		var req map[string]any
		_ = json.Unmarshal(b, &req)
		if req["stream"] == true {
			t.Fatalf("expected stream=false")
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"model":"m1","choices":[{"message":{"role":"assistant","content":"hi"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":2,"total_tokens":3}}`)
	}))
	defer srv.Close()

	client := data.OpenAIClient{BaseURL: srv.URL, DefaultModel: "m1"}
	resp, err := client.Chat(context.Background(), domain.ChatRequest{Messages: []domain.Message{{Role: domain.RoleUser, Content: "hello"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "hi" || resp.Model != "m1" || resp.Provider != "openai" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Usage.TotalTokens != 3 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestOpenAIClient_StreamChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"Hel\"},\"finish_reason\":null}]}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		if ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	client := data.OpenAIClient{BaseURL: srv.URL, DefaultModel: "m1"}

	var deltas []string
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "m1", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(deltas, "") != "Hello" {
		t.Fatalf("unexpected deltas: %v", deltas)
	}
	if resp.Content != "Hello" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if resp.FinishReason != "stop" {
		t.Fatalf("unexpected finish reason: %q", resp.FinishReason)
	}
	if resp.Model != "m1" {
		t.Fatalf("unexpected model: %q", resp.Model)
	}
}

func TestOpenAIClient_StatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, "rate limit")
	}))
	defer srv.Close()

	client := data.OpenAIClient{BaseURL: srv.URL}
	_, err := client.Chat(context.Background(), domain.ChatRequest{Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}})
	if err == nil || !errors.Is(err, domain.ErrProvider) {
		t.Fatalf("expected ErrProvider, got %v", err)
	}
}

func TestOpenAIClient_StreamInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, "data: {not-json}\n\n")
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	client := data.OpenAIClient{BaseURL: srv.URL, DefaultModel: "m1"}
	_, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "m1", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error { return nil })
	if err == nil || !errors.Is(err, domain.ErrStream) {
		t.Fatalf("expected ErrStream, got %v", err)
	}
}

type mockDoer struct {
	do func(req *http.Request) (*http.Response, error)
}

func (m mockDoer) Do(req *http.Request) (*http.Response, error) {
	return m.do(req)
}

func newResponse(statusCode int, body string, contentType string) *http.Response {
	h := make(http.Header)
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d", statusCode),
		Header:     h,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestClaudeClient_Chat_UsesMockHTTPClient(t *testing.T) {
	var gotPath string
	var gotSystem string

	client := data.ClaudeClient{
		BaseURL: "https://example.invalid",
		APIKey:  "k",
		HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
			gotPath = req.URL.Path
			if req.Header.Get("anthropic-version") == "" {
				t.Fatalf("expected anthropic-version header")
			}
			b, _ := io.ReadAll(req.Body)
			_ = req.Body.Close()
			var payload map[string]any
			_ = json.Unmarshal(b, &payload)
			if s, ok := payload["system"].(string); ok {
				gotSystem = s
			}
			return newResponse(200, `{"model":"claude-m","stop_reason":"end_turn","content":[{"type":"text","text":"Hi"}],"usage":{"input_tokens":2,"output_tokens":3}}`, "application/json"), nil
		}},
	}

	resp, err := client.Chat(context.Background(), domain.ChatRequest{Messages: []domain.Message{{Role: domain.RoleSystem, Content: "sys"}, {Role: domain.RoleUser, Content: "u"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/v1/messages" {
		t.Fatalf("unexpected path: %q", gotPath)
	}
	if gotSystem != "sys" {
		t.Fatalf("unexpected system: %q", gotSystem)
	}
	if resp.Provider != "claude" || resp.Content != "Hi" || resp.FinishReason != "end_turn" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestClaudeClient_StreamChat_ParsesSSE(t *testing.T) {
	sse := strings.Join([]string{
		"event: message_start",
		"data: {\"message\":{\"model\":\"claude-x\"}}",
		"",
		"event: content_block_delta",
		"data: {\"delta\":{\"text\":\"Hel\"}}",
		"",
		"event: content_block_delta",
		"data: {\"delta\":{\"text\":\"lo\"}}",
		"",
		"event: message_delta",
		"data: {\"delta\":{\"stop_reason\":\"end_turn\"}}",
		"",
		"event: message_stop",
		"data: {}",
		"",
	}, "\n")

	client := data.ClaudeClient{BaseURL: "https://example.invalid", HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
		return newResponse(200, sse, "text/event-stream"), nil
	}}}

	var deltas []string
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "m", Messages: []domain.Message{{Role: domain.RoleUser, Content: "u"}}}, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Join(deltas, "") != "Hello" {
		t.Fatalf("unexpected deltas: %v", deltas)
	}
	if resp.Content != "Hello" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if resp.FinishReason != "end_turn" {
		t.Fatalf("unexpected finish reason: %q", resp.FinishReason)
	}
	if resp.Model != "claude-x" {
		t.Fatalf("unexpected model: %q", resp.Model)
	}
}

func TestGeminiClient_Chat_UsesAPIKeyQueryParam(t *testing.T) {
	var gotURL *url.URL
	client := data.GeminiClient{BaseURL: "https://example.invalid", APIKey: "abc", HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
		gotURL = req.URL
		return newResponse(200, `{"candidates":[{"content":{"parts":[{"text":"OK"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":2,"totalTokenCount":3}}`, "application/json"), nil
	}}}

	resp, err := client.Chat(context.Background(), domain.ChatRequest{Model: "gemini-1.5-flash", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotURL == nil || gotURL.Query().Get("key") != "abc" {
		t.Fatalf("expected key query param, got %v", gotURL)
	}
	if resp.Provider != "gemini" || resp.Content != "OK" || resp.FinishReason != "STOP" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Usage.TotalTokens != 3 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestGeminiClient_StreamChat_ParsesSSE(t *testing.T) {
	sse := strings.Join([]string{
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"A\"}]} }]}",
		"",
		"data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"B\"}]},\"finishReason\":\"STOP\"}]}",
		"",
	}, "\n")

	client := data.GeminiClient{BaseURL: "https://example.invalid", HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
		return newResponse(200, sse, "text/event-stream"), nil
	}}}

	var got string
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "gemini-1.5-flash", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error {
		got += delta
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "AB" {
		t.Fatalf("unexpected deltas: %q", got)
	}
	if resp.Content != "AB" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if resp.FinishReason != "STOP" {
		t.Fatalf("unexpected finish: %q", resp.FinishReason)
	}
}

func TestClaudeClient_StatusError(t *testing.T) {
	client := data.ClaudeClient{BaseURL: "https://example.invalid", HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
		return newResponse(429, "rate limit", "text/plain"), nil
	}}}
	_, err := client.Chat(context.Background(), domain.ChatRequest{Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}})
	if err == nil || !errors.Is(err, domain.ErrProvider) {
		t.Fatalf("expected ErrProvider, got %v", err)
	}
}

func TestGeminiClient_StreamInvalidJSON(t *testing.T) {
	client := data.GeminiClient{BaseURL: "https://example.invalid", HTTPClient: mockDoer{do: func(req *http.Request) (*http.Response, error) {
		return newResponse(200, "data: {not-json}\n\n", "text/event-stream"), nil
	}}}
	_, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "gemini-1.5-flash", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error { return nil })
	if err == nil || !errors.Is(err, domain.ErrStream) {
		t.Fatalf("expected ErrStream, got %v", err)
	}
}
