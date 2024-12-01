package interfaces

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/gs1068/slack-gpt-bot/usecase"
)

type GptHandler struct {
	gptUsecase *usecase.GptUsecase
}

func NewGptHandler(gptUsecase *usecase.GptUsecase) GptHandler {
	return GptHandler{
		gptUsecase: gptUsecase,
	}
}

func (h *GptHandler) CreateCompletion(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		log.Fatal().Msg("failed r.URL.Query().Get(\"prompt\")")
		http.Error(w, "failed r.URL.Query().Get(\"prompt\")", http.StatusBadRequest)
		return
	}

	resp, err := h.gptUsecase.CreateCompletion(ctx, prompt)
	if err != nil {
		log.Fatal().Err(err).Msg("failed h.gptUsecase.CreateCompletion")
		http.Error(w, "failed to create completion: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *GptHandler) CreateImage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		log.Error().Msg("missing prompt parameter")
		http.Error(w, "missing prompt parameter", http.StatusBadRequest)
		return
	}

	respImage, err := h.gptUsecase.CreateImage(ctx, prompt)
	if err != nil {
		log.Error().Err(err).Msg("failed to create image")
		http.Error(w, "failed to create image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, respImage)
	if err != nil {
		log.Error().Err(err).Msg("failed to write image data to response")
		http.Error(w, "failed to send image data", http.StatusInternalServerError)
		return
	}
}
