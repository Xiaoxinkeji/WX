package aiwriting_test

import (
	"context"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/usecase"
)

type providerStub struct{ content string }

func (p providerStub) ProviderName() string { return "stub" }

func (p providerStub) Chat(ctx context.Context, req domain.ChatRequest) (domain.ChatResponse, error) {
	return domain.ChatResponse{Provider: p.ProviderName(), Model: req.Model, Content: p.content}, nil
}

func (p providerStub) StreamChat(ctx context.Context, req domain.ChatRequest, onDelta func(delta string) error) (domain.ChatResponse, error) {
	if err := onDelta(p.content); err != nil {
		return domain.ChatResponse{}, err
	}
	return domain.ChatResponse{Provider: p.ProviderName(), Model: req.Model, Content: p.content, FinishReason: "stop"}, nil
}

func TestGenerateContentUseCase_ReturnsGeneration(t *testing.T) {
	prompts := data.NewDefaultPromptRepository()
	uc := usecase.NewGenerateContentUseCase(prompts, providerStub{content: "hello"})
	out, err := uc.Execute(context.Background(), usecase.GenerateContentInput{Topic: "Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Generation.Type != domain.GenerationTypeGenerate {
		t.Fatalf("unexpected type: %q", out.Generation.Type)
	}
	if out.Generation.OutputText != "hello" {
		t.Fatalf("unexpected output: %q", out.Generation.OutputText)
	}
	if out.Article != nil {
		t.Fatalf("expected no article, got %+v", out.Article)
	}
}

func TestRewriteAndSummarizeUseCases_Work(t *testing.T) {
	prompts := data.NewDefaultPromptRepository()
	provider := providerStub{content: "ok"}

	rewriteUC := usecase.NewRewriteContentUseCase(prompts, provider)
	rewriteOut, err := rewriteUC.Execute(context.Background(), usecase.RewriteContentInput{Text: "hello"})
	if err != nil {
		t.Fatalf("rewrite: %v", err)
	}
	if rewriteOut.Generation.Type != domain.GenerationTypeRewrite {
		t.Fatalf("unexpected rewrite type: %q", rewriteOut.Generation.Type)
	}

	sumUC := usecase.NewSummarizeUseCase(prompts, provider)
	sumOut, err := sumUC.Execute(context.Background(), usecase.SummarizeInput{Text: "hello"})
	if err != nil {
		t.Fatalf("summarize: %v", err)
	}
	if sumOut.Generation.Type != domain.GenerationTypeSummarize {
		t.Fatalf("unexpected summarize type: %q", sumOut.Generation.Type)
	}
}
