package interfaces

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gs1068/slack-gpt-bot/usecase"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack/slackevents"
)

type SlackHandler struct {
	slackUsecase *usecase.SlackUsecase
}

func NewSlackHandler(slackUsecase *usecase.SlackUsecase) SlackHandler {
	return SlackHandler{
		slackUsecase: slackUsecase,
	}
}

func (i *SlackHandler) EventHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Slack APPはレスポンスが遅かったりするとリトライが行われる。
	// GPT側でリクエストを重複して処理してしまうのを防ぐため、リトライの場合は無視する。
	if retryNum := r.Header.Get("X-Slack-Retry-Num"); retryNum != "" {
		log.Info().Msg("retry request")
		w.WriteHeader(http.StatusOK)
		return
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal().Msg("failed io.ReadAll(r.Body)")
		httpError(w, "failed to read request body", http.StatusInternalServerError, err)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(bodyBytes), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Fatal().Err(err).Msg("failed slackevents.ParseEvent")
		httpError(w, "invalid event", http.StatusInternalServerError, err)
		return
	}

	switch event := eventsAPIEvent.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		handleAppMentionEvent(ctx, w, i.slackUsecase, event)
	case *slackevents.MessageEvent:
		handleMessageEvent(ctx, w, i.slackUsecase, event)
	default:
		log.Info().Msg("unsupported event")
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleAppMentionEvent(ctx context.Context, w http.ResponseWriter, usecase *usecase.SlackUsecase, event *slackevents.AppMentionEvent) {
	ts := getThreadTimestamp(event.TimeStamp, event.ThreadTimeStamp)

	if err := usecase.ProcessMessages(ctx, event.Channel, ts); err != nil {
		log.Fatal().Err(err).Msg("failed usecase.ProcessMessages")
		httpError(w, "failed to process mention event", http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleMessageEvent(ctx context.Context, w http.ResponseWriter, usecase *usecase.SlackUsecase, event *slackevents.MessageEvent) {
	if event.User == "" || event.BotID != "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ts := getThreadTimestamp(event.TimeStamp, event.ThreadTimeStamp)

	switch {
	case event.ChannelType == "im":
		if err := usecase.ProcessMessages(ctx, event.Channel, ts); err != nil {
			log.Fatal().Err(err).Msg("failed usecase.ProcessMessages")
			httpError(w, "failed to process direct message", http.StatusInternalServerError, err)
			return
		}
	case event.ThreadTimeStamp != "":
		if err := usecase.ProcessMessages(ctx, event.Channel, ts); err != nil {
			log.Fatal().Err(err).Msg("failed usecase.ProcessMessages")
			httpError(w, "failed to process thread messages", http.StatusInternalServerError, err)
			return
		}
	default:
		log.Info().Msg("unsupported message event")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getThreadTimestamp(timeStamp, threadTimeStamp string) string {
	if threadTimeStamp != "" {
		return threadTimeStamp
	}
	return timeStamp
}

func httpError(w http.ResponseWriter, message string, statusCode int, err error) {
	log.Error().Err(err).Msg(message)
	http.Error(w, message, statusCode)
}
