package models

import (
	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type PromptMessageDTO struct {
	Role     string `json:"role"`
	Template string `json:"template"`
}

type PromptDTO struct {
	Name     string            `json:"name"`
	Messages []PromptMessageDTO `json:"messages"`
}

func FromPrompt(p domain.Prompt) PromptDTO {
	msgs := make([]PromptMessageDTO, 0, len(p.Messages))
	for _, m := range p.Messages {
		msgs = append(msgs, PromptMessageDTO{Role: string(m.Role), Template: m.Template})
	}
	return PromptDTO{Name: p.Name, Messages: msgs}
}

func (d PromptDTO) ToDomain() (domain.Prompt, error) {
	msgs := make([]domain.PromptMessage, 0, len(d.Messages))
	for _, m := range d.Messages {
		msgs = append(msgs, domain.PromptMessage{Role: domain.Role(m.Role), Template: m.Template})
	}
	p := domain.Prompt{Name: d.Name, Messages: msgs}
	return p, p.Validate()
}
