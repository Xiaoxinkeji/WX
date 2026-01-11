package domain

import "time"

type GenerationType string

const (
	GenerationTypeGenerate   GenerationType = "generate"
	GenerationTypeRewrite    GenerationType = "rewrite"
	GenerationTypeSummarize  GenerationType = "summarize"
	GenerationTypeContinuous GenerationType = "continue"
)

type Generation struct {
	ID         string
	Type       GenerationType
	PromptName string
	Provider   string
	Model      string
	InputText  string
	OutputText string
	ArticleID  string
	CreatedAt  time.Time
}

type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
	MaxTokens   int
}

type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

type ChatResponse struct {
	Provider     string
	Model        string
	Content      string
	FinishReason string
	Usage        TokenUsage
}
