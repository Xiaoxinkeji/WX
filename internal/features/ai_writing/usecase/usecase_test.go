package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	ai "github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/usecase"
	articles "github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type seqIDs struct {
	ids []string
	i   int
}

func (g *seqIDs) NewID() (string, error) {
	if g.i >= len(g.ids) {
		return "", errors.New("no more ids")
	}
	id := g.ids[g.i]
	g.i++
	return id, nil
}

type promptRepoFake struct {
	prompt ai.Prompt
	err    error
	name   string
}

func (f *promptRepoFake) GetPrompt(ctx context.Context, name string) (ai.Prompt, error) {
	f.name = name
	return f.prompt, f.err
}

type providerFake struct {
	calledChat   int
	calledStream int

	content string
	deltas  []string
}

func (p *providerFake) ProviderName() string { return "fake" }

func (p *providerFake) Chat(ctx context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	p.calledChat++
	return ai.ChatResponse{Provider: p.ProviderName(), Model: req.Model, Content: p.content}, nil
}

func (p *providerFake) StreamChat(ctx context.Context, req ai.ChatRequest, onDelta func(delta string) error) (ai.ChatResponse, error) {
	p.calledStream++
	for _, d := range p.deltas {
		if err := onDelta(d); err != nil {
			return ai.ChatResponse{}, err
		}
	}
	return ai.ChatResponse{Provider: p.ProviderName(), Model: req.Model, Content: p.content, FinishReason: "stop"}, nil
}

type articleCreatorFake struct {
	called int
	params articles.CreateArticleParams
}

func (f *articleCreatorFake) CreateArticle(ctx context.Context, params articles.CreateArticleParams) (articles.Article, error) {
	f.called++
	f.params = params
	return articles.Article{ID: params.ID, Title: params.Title, Content: params.Content, Status: params.Status, CreatedAt: params.CreatedAt, UpdatedAt: params.UpdatedAt}, nil
}

func TestGenerateContentUseCase_ValidatesTopic(t *testing.T) {
	uc := usecase.GenerateContentUseCase{Prompts: &promptRepoFake{}, Provider: &providerFake{}}
	_, err := uc.Execute(context.Background(), usecase.GenerateContentInput{Topic: "  "})
	if err == nil || !errors.Is(err, ai.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}

func TestGenerateContentUseCase_CallsProviderAndSavesDraft(t *testing.T) {
	now := time.Date(2025, 4, 5, 6, 7, 8, 0, time.UTC)
	ids := &seqIDs{ids: []string{"gen-1", "art-1"}}
	provider := &providerFake{content: "generated"}
	articlesRepo := &articleCreatorFake{}
	repo := &promptRepoFake{prompt: ai.Prompt{Name: "generate_content", Messages: []ai.PromptMessage{{Role: ai.RoleUser, Template: "topic={{topic}}"}}}}

	uc := usecase.GenerateContentUseCase{Prompts: repo, Provider: provider, Articles: articlesRepo, Clock: fixedClock{t: now}, IDs: ids}
	out, err := uc.Execute(context.Background(), usecase.GenerateContentInput{Topic: "T", SaveAsDraft: true, Tags: []string{" Go ", "go"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.calledChat != 1 {
		t.Fatalf("expected provider chat called once, got %d", provider.calledChat)
	}
	if articlesRepo.called != 1 {
		t.Fatalf("expected articles CreateArticle called once, got %d", articlesRepo.called)
	}
	if articlesRepo.params.Title != "T" {
		t.Fatalf("expected title T, got %q", articlesRepo.params.Title)
	}
	if !reflect.DeepEqual(articlesRepo.params.Tags, []string{"go"}) {
		t.Fatalf("unexpected tags: %v", articlesRepo.params.Tags)
	}
	if out.Generation.ID != "gen-1" {
		t.Fatalf("expected gen-1, got %q", out.Generation.ID)
	}
	if out.Generation.ArticleID != "art-1" {
		t.Fatalf("expected art-1, got %q", out.Generation.ArticleID)
	}
	if out.Article == nil || out.Article.ID != "art-1" {
		t.Fatalf("unexpected article: %+v", out.Article)
	}
}

func TestGenerateContentUseCase_StreamingPassesDeltas(t *testing.T) {
	provider := &providerFake{content: "ab", deltas: []string{"a", "b"}}
	repo := &promptRepoFake{prompt: ai.Prompt{Name: "generate_content", Messages: []ai.PromptMessage{{Role: ai.RoleUser, Template: "topic={{topic}}"}}}}
	uc := usecase.GenerateContentUseCase{Prompts: repo, Provider: provider, Clock: fixedClock{t: time.Unix(0, 0).UTC()}, IDs: &seqIDs{ids: []string{"gen-1"}}}

	var got []string
	out, err := uc.Execute(context.Background(), usecase.GenerateContentInput{Topic: "T", OnDelta: func(delta string) error { got = append(got, delta); return nil }})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.calledStream != 1 {
		t.Fatalf("expected provider stream called once, got %d", provider.calledStream)
	}
	if !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Fatalf("unexpected deltas: %v", got)
	}
	if out.Generation.OutputText != "ab" {
		t.Fatalf("unexpected content: %q", out.Generation.OutputText)
	}
}

func TestRewriteAndSummarize_RequireText(t *testing.T) {
	provider := &providerFake{content: "x"}
	repo := &promptRepoFake{prompt: ai.Prompt{Name: "rewrite_content", Messages: []ai.PromptMessage{{Role: ai.RoleUser, Template: "{{text}}"}}}}

	rewriteUC := usecase.RewriteContentUseCase{Prompts: repo, Provider: provider, Clock: fixedClock{t: time.Unix(0, 0).UTC()}, IDs: &seqIDs{ids: []string{"id-1"}}}
	_, err := rewriteUC.Execute(context.Background(), usecase.RewriteContentInput{Text: " "})
	if err == nil || !errors.Is(err, ai.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}

	sumRepo := &promptRepoFake{prompt: ai.Prompt{Name: "summarize", Messages: []ai.PromptMessage{{Role: ai.RoleUser, Template: "{{text}}"}}}}
	sumUC := usecase.SummarizeUseCase{Prompts: sumRepo, Provider: provider, Clock: fixedClock{t: time.Unix(0, 0).UTC()}, IDs: &seqIDs{ids: []string{"id-2"}}}
	_, err = sumUC.Execute(context.Background(), usecase.SummarizeInput{Text: ""})
	if err == nil || !errors.Is(err, ai.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
