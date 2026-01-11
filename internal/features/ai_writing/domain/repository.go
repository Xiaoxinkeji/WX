package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("ai_writing: not found")
	ErrInvalidArgument = errors.New("ai_writing: invalid argument")
	ErrProvider        = errors.New("ai_writing: provider error")
	ErrStream          = errors.New("ai_writing: stream error")
)

type PromptRepository interface {
	GetPrompt(ctx context.Context, name string) (Prompt, error)
}

type Clock interface {
	Now() time.Time
}

type IDGenerator interface {
	NewID() (string, error)
}

type Provider interface {
	ProviderName() string
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	StreamChat(ctx context.Context, req ChatRequest, onDelta func(delta string) error) (ChatResponse, error)
}
