package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type SummarizeInput struct {
	Text        string
	PromptName  string
	Model       string
	Temperature float64
	MaxTokens   int
	OnDelta     func(delta string) error
}

type SummarizeOutput struct {
	Generation domain.Generation
}

type SummarizeUseCase struct {
	Prompts  domain.PromptRepository
	Provider domain.Provider
	Clock    domain.Clock
	IDs      domain.IDGenerator
}

func NewSummarizeUseCase(prompts domain.PromptRepository, provider domain.Provider) SummarizeUseCase {
	return SummarizeUseCase{Prompts: prompts, Provider: provider, Clock: systemClock{}, IDs: randomIDGenerator{}}
}

func (uc SummarizeUseCase) Execute(ctx context.Context, in SummarizeInput) (SummarizeOutput, error) {
	if uc.Prompts == nil {
		return SummarizeOutput{}, errors.New("summarize: prompts is nil")
	}
	if uc.Provider == nil {
		return SummarizeOutput{}, errors.New("summarize: provider is nil")
	}
	if uc.Clock == nil {
		uc.Clock = systemClock{}
	}
	if uc.IDs == nil {
		uc.IDs = randomIDGenerator{}
	}

	text := strings.TrimSpace(in.Text)
	if text == "" {
		return SummarizeOutput{}, errors.Join(domain.ErrInvalidArgument, errors.New("text is required"))
	}
	promptName := strings.TrimSpace(in.PromptName)
	if promptName == "" {
		promptName = "summarize"
	}

	prompt, err := uc.Prompts.GetPrompt(ctx, promptName)
	if err != nil {
		return SummarizeOutput{}, err
	}
	msgs, err := prompt.Render(map[string]string{"text": text})
	if err != nil {
		return SummarizeOutput{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	chatReq := domain.ChatRequest{Model: in.Model, Messages: msgs, Temperature: in.Temperature, MaxTokens: in.MaxTokens}
	var chatResp domain.ChatResponse
	if in.OnDelta == nil {
		chatResp, err = uc.Provider.Chat(ctx, chatReq)
	} else {
		chatResp, err = uc.Provider.StreamChat(ctx, chatReq, in.OnDelta)
	}
	if err != nil {
		return SummarizeOutput{}, err
	}

	id, err := uc.IDs.NewID()
	if err != nil {
		return SummarizeOutput{}, err
	}
	gen := domain.Generation{
		ID:         id,
		Type:       domain.GenerationTypeSummarize,
		PromptName: promptName,
		Provider:   chatResp.Provider,
		Model:      chatResp.Model,
		InputText:  text,
		OutputText: chatResp.Content,
		CreatedAt:  uc.Clock.Now(),
	}
	return SummarizeOutput{Generation: gen}, nil
}
