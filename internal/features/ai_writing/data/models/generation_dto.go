package models

import (
	"time"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type GenerationDTO struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	PromptName string `json:"prompt_name"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	InputText  string `json:"input_text"`
	OutputText string `json:"output_text"`
	ArticleID  string `json:"article_id"`
	CreatedAtMs int64 `json:"created_at_ms"`
}

func FromGeneration(g domain.Generation) GenerationDTO {
	createdAtMs := g.CreatedAt.UTC().UnixMilli()
	return GenerationDTO{
		ID:          g.ID,
		Type:        string(g.Type),
		PromptName:  g.PromptName,
		Provider:    g.Provider,
		Model:       g.Model,
		InputText:   g.InputText,
		OutputText:  g.OutputText,
		ArticleID:   g.ArticleID,
		CreatedAtMs: createdAtMs,
	}
}

func (d GenerationDTO) ToDomain() domain.Generation {
	created := time.UnixMilli(d.CreatedAtMs).UTC()
	return domain.Generation{
		ID:         d.ID,
		Type:       domain.GenerationType(d.Type),
		PromptName: d.PromptName,
		Provider:   d.Provider,
		Model:      d.Model,
		InputText:  d.InputText,
		OutputText: d.OutputText,
		ArticleID:  d.ArticleID,
		CreatedAt:  created,
	}
}
