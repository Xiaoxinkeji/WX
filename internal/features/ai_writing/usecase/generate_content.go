package usecase

import (
	"context"
	"errors"
	"strings"

	ai "github.com/Xiaoxinkeji/WX/internal/features/ai_writing/domain"
	articles "github.com/Xiaoxinkeji/WX/internal/features/articles/domain"
)

type GenerateContentInput struct {
	Topic       string
	PromptName  string
	Model       string
	Temperature float64
	MaxTokens   int
	OnDelta     func(delta string) error

	SaveAsDraft bool
	DraftTitle  string
	Tags        []string
}

type GenerateContentOutput struct {
	Generation ai.Generation
	Article    *articles.Article
}

type GenerateContentUseCase struct {
	Prompts  ai.PromptRepository
	Provider ai.Provider
	Articles articles.ArticleCreator
	Clock    ai.Clock
	IDs      ai.IDGenerator
}

func NewGenerateContentUseCase(prompts ai.PromptRepository, provider ai.Provider) GenerateContentUseCase {
	return GenerateContentUseCase{Prompts: prompts, Provider: provider, Clock: systemClock{}, IDs: randomIDGenerator{}}
}

func (uc GenerateContentUseCase) Execute(ctx context.Context, in GenerateContentInput) (GenerateContentOutput, error) {
	if uc.Prompts == nil {
		return GenerateContentOutput{}, errors.New("generate content: prompts is nil")
	}
	if uc.Provider == nil {
		return GenerateContentOutput{}, errors.New("generate content: provider is nil")
	}
	if uc.Clock == nil {
		uc.Clock = systemClock{}
	}
	if uc.IDs == nil {
		uc.IDs = randomIDGenerator{}
	}

	topic := strings.TrimSpace(in.Topic)
	if topic == "" {
		return GenerateContentOutput{}, errors.Join(ai.ErrInvalidArgument, errors.New("topic is required"))
	}
	promptName := strings.TrimSpace(in.PromptName)
	if promptName == "" {
		promptName = "generate_content"
	}

	prompt, err := uc.Prompts.GetPrompt(ctx, promptName)
	if err != nil {
		return GenerateContentOutput{}, err
	}
	msgs, err := prompt.Render(map[string]string{"topic": topic})
	if err != nil {
		return GenerateContentOutput{}, errors.Join(ai.ErrInvalidArgument, err)
	}

	chatReq := ai.ChatRequest{Model: in.Model, Messages: msgs, Temperature: in.Temperature, MaxTokens: in.MaxTokens}
	var chatResp ai.ChatResponse
	if in.OnDelta == nil {
		chatResp, err = uc.Provider.Chat(ctx, chatReq)
	} else {
		chatResp, err = uc.Provider.StreamChat(ctx, chatReq, in.OnDelta)
	}
	if err != nil {
		return GenerateContentOutput{}, err
	}

	genID, err := uc.IDs.NewID()
	if err != nil {
		return GenerateContentOutput{}, err
	}

	now := uc.Clock.Now()
	gen := ai.Generation{
		ID:         genID,
		Type:       ai.GenerationTypeGenerate,
		PromptName: promptName,
		Provider:   chatResp.Provider,
		Model:      chatResp.Model,
		InputText:  topic,
		OutputText: chatResp.Content,
		CreatedAt:  now,
	}

	var createdArticle *articles.Article
	if in.SaveAsDraft {
		if uc.Articles == nil {
			return GenerateContentOutput{}, errors.New("generate content: articles repo is nil")
		}
		title := strings.TrimSpace(in.DraftTitle)
		if title == "" {
			title = topic
		}
		normalizedTags, err := articles.NormalizeTagNames(in.Tags)
		if err != nil {
			return GenerateContentOutput{}, errors.Join(articles.ErrInvalidArgument, err)
		}

		articleID, err := uc.IDs.NewID()
		if err != nil {
			return GenerateContentOutput{}, err
		}
		article, err := uc.Articles.CreateArticle(ctx, articles.CreateArticleParams{
			ID:        articleID,
			Title:     title,
			Content:   chatResp.Content,
			Status:    articles.ArticleStatusDraft,
			Tags:      normalizedTags,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			return GenerateContentOutput{}, err
		}
		createdArticle = &article
		gen.ArticleID = article.ID
	}

	return GenerateContentOutput{Generation: gen, Article: createdArticle}, nil
}
