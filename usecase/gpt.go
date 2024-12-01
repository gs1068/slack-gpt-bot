package usecase

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gs1068/slack-gpt-bot/domain/repository"
	"github.com/sashabaranov/go-openai"
)

type GptUsecase struct {
	gpt repository.GptRepository
}

func NewGptUsecase(
	gpt repository.GptRepository,
) *GptUsecase {
	return &GptUsecase{
		gpt: gpt,
	}
}

func (u *GptUsecase) CreateCompletion(ctx context.Context, prompt string) (openai.ChatCompletionResponse, error) {
	resp, err := u.gpt.CreateCompletion(ctx, prompt)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("failed u.gpt.CreateCompletion: %w", err)
	}

	return resp, nil
}

func (u *GptUsecase) CreateImage(ctx context.Context, prompt string) (io.Reader, error) {
	respUrl, err := u.gpt.CreateImage(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed u.gpt.CreateImage: %w", err)
	}

	image, err := downloadImage(respUrl)
	if err != nil {
		return nil, fmt.Errorf("failed downloadImage: %w", err)
	}

	return image, nil
}

func downloadImage(url string) (io.Reader, error) {
	response, err := http.Get(url)
	if err != nil {
		log.Printf("downloadImage failed, err=%+v", err)
		return nil, err
	}
	return response.Body, nil
}
