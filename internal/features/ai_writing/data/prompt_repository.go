package data

import (
	"context"
	"errors"
	"strings"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type InMemoryPromptRepository struct {
	prompts map[string]domain.Prompt
}

func NewInMemoryPromptRepository(prompts []domain.Prompt) (*InMemoryPromptRepository, error) {
	repo := &InMemoryPromptRepository{prompts: map[string]domain.Prompt{}}
	for _, p := range prompts {
		if err := p.Validate(); err != nil {
			return nil, err
		}
		repo.prompts[strings.ToLower(strings.TrimSpace(p.Name))] = p
	}
	return repo, nil
}

func NewDefaultPromptRepository() *InMemoryPromptRepository {
	repo, _ := NewInMemoryPromptRepository([]domain.Prompt{
		{
			Name: "generate_content",
			Messages: []domain.PromptMessage{
				{Role: domain.RoleSystem, Template: "You are a helpful writing assistant."},
				{Role: domain.RoleUser, Template: "Write a well-structured WeChat-style article about: {{topic}}"},
			},
		},
		{
			Name: "rewrite_content",
			Messages: []domain.PromptMessage{
				{Role: domain.RoleSystem, Template: "You are a helpful writing assistant."},
				{Role: domain.RoleUser, Template: "Rewrite the following text to improve clarity and style, keeping meaning unchanged:\n\n{{text}}"},
			},
		},
		{
			Name: "summarize",
			Messages: []domain.PromptMessage{
				{Role: domain.RoleSystem, Template: "You are a helpful writing assistant."},
				{Role: domain.RoleUser, Template: "Summarize the following text into concise bullet points:\n\n{{text}}"},
			},
		},
	})
	if repo == nil {
		return &InMemoryPromptRepository{prompts: map[string]domain.Prompt{}}
	}
	return repo
}

func (r *InMemoryPromptRepository) GetPrompt(ctx context.Context, name string) (domain.Prompt, error) {
	_ = ctx
	if r == nil {
		return domain.Prompt{}, errors.New("prompt repository: nil")
	}
	key := strings.ToLower(strings.TrimSpace(name))
	if key == "" {
		return domain.Prompt{}, errors.Join(domain.ErrInvalidArgument, errors.New("prompt name is required"))
	}
	p, ok := r.prompts[key]
	if !ok {
		return domain.Prompt{}, errors.Join(domain.ErrNotFound, errors.New("prompt not found"))
	}
	return p, nil
}
