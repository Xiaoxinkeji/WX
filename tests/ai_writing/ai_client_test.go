package aiwriting_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

func TestAIClient_StreamChat_ParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"model\":\"m\",\"choices\":[{\"delta\":{\"content\":\"A\"}}]}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "data: {\"model\":\"m\",\"choices\":[{\"delta\":{\"content\":\"B\"},\"finish_reason\":\"stop\"}]}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		if ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	client := data.OpenAIClient{BaseURL: srv.URL, DefaultModel: "m"}
	var got string
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "m", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error {
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
}

func TestClaudeClient_StreamChat_ParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		fmt.Fprint(w, "event: message_start\n")
		fmt.Fprint(w, "data: {\"message\":{\"model\":\"claude-test\"}}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "event: content_block_delta\n")
		fmt.Fprint(w, "data: {\"delta\":{\"text\":\"Hi\"}}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "event: message_delta\n")
		fmt.Fprint(w, "data: {\"delta\":{\"stop_reason\":\"end_turn\"}}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "event: message_stop\n")
		fmt.Fprint(w, "data: {}\n\n")
		if ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	client := data.ClaudeClient{BaseURL: srv.URL}
	var got strings.Builder
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Messages: []domain.Message{{Role: domain.RoleUser, Content: "x"}}}, func(delta string) error {
		got.WriteString(delta)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "Hi" {
		t.Fatalf("unexpected deltas: %q", got.String())
	}
	if resp.Model != "claude-test" {
		t.Fatalf("unexpected model: %q", resp.Model)
	}
	if resp.FinishReason != "end_turn" {
		t.Fatalf("unexpected finish: %q", resp.FinishReason)
	}
}

func TestGeminiClient_Chat_ParsesResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/v1beta/models/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"candidates":[{"content":{"parts":[{"text":"Hello"}]},"finishReason":"STOP"}],"usageMetadata":{"promptTokenCount":1,"candidatesTokenCount":2,"totalTokenCount":3}}`)
	}))
	defer srv.Close()

	client := data.GeminiClient{BaseURL: srv.URL}
	resp, err := client.Chat(context.Background(), domain.ChatRequest{Model: "gemini-1.5-flash", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "Hello" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if resp.Usage.TotalTokens != 3 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestGeminiClient_StreamChat_ParsesSSE(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"A\"}]}}]}\n\n")
		if ok {
			flusher.Flush()
		}
		fmt.Fprint(w, "data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"B\"}]},\"finishReason\":\"STOP\"}]}\n\n")
		if ok {
			flusher.Flush()
		}
	}))
	defer srv.Close()

	client := data.GeminiClient{BaseURL: srv.URL}
	var got strings.Builder
	resp, err := client.StreamChat(context.Background(), domain.ChatRequest{Model: "gemini-1.5-flash", Messages: []domain.Message{{Role: domain.RoleUser, Content: "hi"}}}, func(delta string) error {
		got.WriteString(delta)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.String() != "AB" {
		t.Fatalf("unexpected deltas: %q", got.String())
	}
	if resp.Content != "AB" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if resp.FinishReason != "STOP" {
		t.Fatalf("unexpected finish: %q", resp.FinishReason)
	}
}
