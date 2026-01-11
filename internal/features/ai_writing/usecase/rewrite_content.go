package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
)

type RewriteContentInput struct {
	Text        string
	PromptName  string
	Model       string
	Temperature float64
	MaxTokens   int
	OnDelta     func(delta string) error
}

type RewriteContentOutput struct {
	Generation domain.Generation
}

type RewriteContentUseCase struct {
	Prompts  domain.PromptRepository
	Provider domain.Provider
	Clock    domain.Clock
	IDs      domain.IDGenerator
}

func NewRewriteContentUseCase(prompts domain.PromptRepository, provider domain.Provider) RewriteContentUseCase {
	return RewriteContentUseCase{Prompts: prompts, Provider: provider, Clock: systemClock{}, IDs: randomIDGenerator{}}
}

func (uc RewriteContentUseCase) Execute(ctx context.Context, in RewriteContentInput) (RewriteContentOutput, error) {
	if uc.Prompts == nil {
		return RewriteContentOutput{}, errors.New("rewrite content: prompts is nil")
	}
	if uc.Provider == nil {
		return RewriteContentOutput{}, errors.New("rewrite content: provider is nil")
	}
	if uc.Clock == nil {
		uc.Clock = systemClock{}
	}
	if uc.IDs == nil {
		uc.IDs = randomIDGenerator{}
	}

	text := strings.TrimSpace(in.Text)
	if text == "" {
		return RewriteContentOutput{}, errors.Join(domain.ErrInvalidArgument, errors.New("text is required"))
	}
	promptName := strings.TrimSpace(in.PromptName)
	if promptName == "" {
		promptName = "rewrite_content"
	}

	prompt, err := uc.Prompts.GetPrompt(ctx, promptName)
	if err != nil {
		return RewriteContentOutput{}, err
	}
	msgs, err := prompt.Render(map[string]string{"text": text})
	if err != nil {
		return RewriteContentOutput{}, errors.Join(domain.ErrInvalidArgument, err)
	}

	chatReq := domain.ChatRequest{Model: in.Model, Messages: msgs, Temperature: in.Temperature, MaxTokens: in.MaxTokens}
	var chatResp domain.ChatResponse
	if in.OnDelta == nil {
		chatResp, err = uc.Provider.Chat(ctx, chatReq)
	} else {
		chatResp, err = uc.Provider.StreamChat(ctx, chatReq, in.OnDelta)
	}
	if err != nil {
		return RewriteContentOutput{}, err
	}

	id, err := uc.IDs.NewID()
	if err != nil {
		return RewriteContentOutput{}, err
	}
	gen := domain.Generation{
		ID:         id,
		Type:       domain.GenerationTypeRewrite,
		PromptName: promptName,
		Provider:   chatResp.Provider,
		Model:      chatResp.Model,
		InputText:  text,
		OutputText: chatResp.Content,
		CreatedAt:  uc.Clock.Now(),
	}
	return RewriteContentOutput{Generation: gen}, nil
}
