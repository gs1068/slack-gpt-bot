package repository

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type GptRepository interface {
	CreateCompletion(ctx context.Context, prompt string) (openai.ChatCompletionResponse, error)
	CreateImage(ctx context.Context, prompt string) (string, error)
}
