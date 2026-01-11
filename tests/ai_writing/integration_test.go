package aiwriting_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/usecase"
	articlesData "github.com/Xiaoxinkeji/WX/internal/features/articles/data"
	articlesDomain "github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("file:ai_writing_suite_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := sql.Open("sqlite", name)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestIntegration_GenerateThenSaveDraft_WithStreaming(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		fmt.Fprint(w, "data: {\"model\":\"m\",\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\n")
		if ok { flusher.Flush() }
		fmt.Fprint(w, "data: {\"model\":\"m\",\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n")
		if ok { flusher.Flush() }
		fmt.Fprint(w, "data: [DONE]\n\n")
		if ok { flusher.Flush() }
	}))
	defer srv.Close()

	db := openTestDB(t)
	articlesRepo, err := articlesData.NewSQLiteRepository(db)
	if err != nil {
		t.Fatalf("new articles repo: %v", err)
	}

	prompts := data.NewDefaultPromptRepository()
	provider := data.OpenAIClient{BaseURL: srv.URL, DefaultModel: "m"}
	uc := usecase.GenerateContentUseCase{Prompts: prompts, Provider: provider, Articles: articlesRepo}

	var deltas []string
	out, err := uc.Execute(context.Background(), usecase.GenerateContentInput{
		Topic:       "My Topic",
		SaveAsDraft: true,
		OnDelta: func(delta string) error {
			deltas = append(deltas, delta)
			return nil
		},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if out.Article == nil {
		t.Fatalf("expected article")
	}
	if out.Article.Content != "Hello" {
		t.Fatalf("unexpected content: %q", out.Article.Content)
	}
	if len(deltas) != 2 {
		t.Fatalf("unexpected deltas: %v", deltas)
	}

	got, err := articlesRepo.GetArticle(context.Background(), out.Article.ID)
	if err != nil {
		t.Fatalf("get article: %v", err)
	}
	if got.Status != articlesDomain.ArticleStatusDraft {
		t.Fatalf("expected draft, got %q", got.Status)
	}
	if got.Content != "Hello" {
		t.Fatalf("unexpected stored content: %q", got.Content)
	}
}

