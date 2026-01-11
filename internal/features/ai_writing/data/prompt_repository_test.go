package data_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/data"
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

func TestDefaultPromptRepository_GetPrompt(t *testing.T) {
	repo := data.NewDefaultPromptRepository()
	p, err := repo.GetPrompt(context.Background(), "generate_content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "generate_content" {
		t.Fatalf("unexpected prompt: %+v", p)
	}
}

func TestDefaultPromptRepository_NotFound(t *testing.T) {
	repo := data.NewDefaultPromptRepository()
	_, err := repo.GetPrompt(context.Background(), "missing")
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDefaultPromptRepository_InvalidName(t *testing.T) {
	repo := data.NewDefaultPromptRepository()
	_, err := repo.GetPrompt(context.Background(), " ")
	if err == nil || !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
}
