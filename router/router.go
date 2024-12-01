package router

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gs1068/slack-gpt-bot/interfaces"
)

func CreateRouter(slackHandler *interfaces.SlackHandler, gptHandler *interfaces.GptHandler) chi.Router {
	r := chi.NewRouter()
	// pingを打つとpongが返ってくるよ
	r.Get("/ping", pingHandler)
	// Slackイベントを受け取るエンドポイント
	r.Post("/events", slackHandler.EventHandler)
	// GPT 検証用なので基本は使わない
	r.Get("/gpt", gptHandler.CreateCompletion)
	r.Get("/gpt/image", gptHandler.CreateImage)

	return r
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, "pong")
}
