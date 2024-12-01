package gpt

import (
	"context"
	"fmt"

	"github.com/gs1068/slack-gpt-bot/domain/model"
	"github.com/gs1068/slack-gpt-bot/domain/repository"
	"github.com/sashabaranov/go-openai"
)

type gptRepository struct {
	gptClient *openai.Client
}

func NewGptRepository(gptClient *openai.Client) repository.GptRepository {
	return &gptRepository{
		gptClient: gptClient,
	}
}

func GptClient(openAiApiKey string) *openai.Client {
	client := openai.NewClient(openAiApiKey)
	return client
}

func (r *gptRepository) CreateCompletion(ctx context.Context, prompt string) (openai.ChatCompletionResponse, error) {
	resp, err := r.gptClient.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleAssistant,
					Content: model.CharacterSettings,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("failed r.gptClient.CreateChatCompletion: %w", err)
	}

	return resp, nil
}

func (r *gptRepository) CreateImage(ctx context.Context, prompt string) (string, error) {
	respUrl, err := r.gptClient.CreateImage(
		ctx,
		openai.ImageRequest{
			Prompt:         prompt,
			Size:           openai.CreateImageSize256x256,
			ResponseFormat: openai.CreateImageResponseFormatURL,
			N:              1,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to create image: %w", err)
	}

	// 生成されたURLを取得
	return respUrl.Data[0].URL, nil
}
