package main

import (
	"context"
	"flag"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gs1068/slack-gpt-bot/config"
	"github.com/gs1068/slack-gpt-bot/infrastructure/gpt"
	"github.com/gs1068/slack-gpt-bot/infrastructure/slack"
	"github.com/gs1068/slack-gpt-bot/infrastructure/spreadsheet"
	"github.com/gs1068/slack-gpt-bot/interfaces"
	"github.com/gs1068/slack-gpt-bot/router"
	"github.com/gs1068/slack-gpt-bot/usecase"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"golang.org/x/sync/errgroup"
)

var (
	logLevel = flag.String("log-level", "info", "Log level")
)

func main() {
	flag.Parse()

	config.LoadEnv()
	slackBotToken := os.Getenv("SLACK_BOT_TOKEN")
	openAIAPIKey := os.Getenv("OPENAI_API_KEY")
	spreadsheet.SpreadsheetID = os.Getenv("SPREADSHEET_ID")

	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.With().Caller().Stack().Logger()

	stdlog.SetFlags(0)
	stdlog.SetOutput(log.Logger)

	lvl, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal().Stack().Err(err).Send()
	}
	zerolog.SetGlobalLevel(lvl)

	// Client
	slackClient := slack.SlackClient(slackBotToken)
	gptClient := gpt.GptClient(openAIAPIKey)
	ssClient, err := spreadsheet.SpreadSheetClient()
	if err != nil {
		log.Fatal().Err(err).Msg("failed spreadsheet.SpreadSheetClient")
	}
	// Repository
	slackRepo := slack.NewSlackRepository(slackClient)
	gptRepo := gpt.NewGptRepository(gptClient)
	ssRepo := spreadsheet.NewSpreadsheetRepository(ssClient)
	// Usecase
	slackUsecase := usecase.NewSlackUsecase(slackRepo, gptRepo, ssRepo)
	gptUsecase := usecase.NewGptUsecase(gptRepo)
	// Handler
	slackHandler := interfaces.NewSlackHandler(slackUsecase)
	gptHandler := interfaces.NewGptHandler(gptUsecase)

	sbu, err := slackRepo.GetBotUserId()
	if err != nil {
		log.Fatal().Err(err).Msg("failed slackRepo.GetBotUserId")
	}
	slack.SlackBotUserID = sbu

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	defer close(sig)

	var g errgroup.Group
	srv := http.Server{
		Addr:    ":8080",
		Handler: router.CreateRouter(&slackHandler, &gptHandler),
	}

	g.Go(func() error {
		log.Info().Str("port", "8080").Msg("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	<-sig
	log.Info().Msg("shutting down server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error().Err(err).Msg("an error occurred while shutting down the server")
	}
	if err := g.Wait(); err != nil {
		log.Error().Err(err).Msg("server error")
	}
}
