package domain_test

import (
	"reflect"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

func TestRenderTemplate_ReplacesVariables(t *testing.T) {
	out, err := domain.RenderTemplate("Hello, {{name}}!", map[string]string{"name": "World"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Hello, World!" {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestRenderTemplate_MissingVariables(t *testing.T) {
	_, err := domain.RenderTemplate("Hello, {{name}}!", map[string]string{})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPrompt_RequiredVariablesAndRender(t *testing.T) {
	p := domain.Prompt{
		Name: "p1",
		Messages: []domain.PromptMessage{
			{Role: domain.RoleSystem, Template: "sys"},
			{Role: domain.RoleUser, Template: "topic={{topic}}; text={{text}}"},
			{Role: domain.RoleUser, Template: "again {{topic}}"},
		},
	}
	vars := p.RequiredVariables()
	if !reflect.DeepEqual(vars, []string{"text", "topic"}) {
		t.Fatalf("unexpected vars: %v", vars)
	}

	msgs, err := p.Render(map[string]string{"topic": "Go", "text": "Hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[1].Content != "topic=Go; text=Hello" {
		t.Fatalf("unexpected content: %q", msgs[1].Content)
	}
}
