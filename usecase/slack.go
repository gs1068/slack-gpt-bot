package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/gs1068/slack-gpt-bot/domain/model"
	"github.com/gs1068/slack-gpt-bot/domain/repository"
	"github.com/gs1068/slack-gpt-bot/infrastructure/slack"
)

type SlackUsecase struct {
	slack repository.SlackRepository
	gpt   repository.GptRepository
	ss    repository.SpreadsheetRepository
}

func NewSlackUsecase(
	slack repository.SlackRepository,
	gpt repository.GptRepository,
	ss repository.SpreadsheetRepository,
) *SlackUsecase {
	return &SlackUsecase{
		slack: slack,
		gpt:   gpt,
		ss:    ss,
	}
}

func (u *SlackUsecase) ProcessMessages(ctx context.Context, channelId string, timeStamp string) error {
	currentData, err := u.ss.GetSpreadsheetDataBySlackID(ctx, slack.SlackBotUserID)
	if err != nil {
		return fmt.Errorf("failed to retrieve spreadsheet data: %w", err)
	}

	// データがない場合はユーザーを新規作成
	if currentData == nil {
		currentData = model.NewSpreadsheet(fmt.Sprint(slack.SlackBotUserID), 0, "", 0, 0, 0)
	}

	// 日付が変わったら使用量をリセット
	currentData.ResetDailyUsageIfNeeded()

	// トークン使用可能かチェック
	if err := currentData.CanUseDailyTokens(); err != nil {
		// 上限を超えた場合はメッセージを返して処理を終了
		return u.slack.CreateNewBotMessage(channelId, timeStamp, model.LimitMessage)
	}

	messages, err := u.slack.LoadConversationReplies(channelId, timeStamp)
	if err != nil {
		return fmt.Errorf("failed u.slack.LoadConversationReplies for channel %s, timestamp %s: %v", channelId, timeStamp, err)
	}

	slackMessages := model.ConvertToSlackMessages(messages)
	botUserID := slack.SlackBotUserID
	gptPrompt := slackMessages.CreatePrompt(botUserID)
	log.Printf("[GPTプロンプト] %s", gptPrompt)

	// GPT応答を取得
	gptResponse, err := u.gpt.CreateCompletion(ctx, gptPrompt)
	if err != nil {
		return fmt.Errorf("failed u.gpt.CreateCompletion: %v", err)
	}

	// GPT応答をメッセージとして追加
	var gptMessage string
	if len(gptResponse.Choices) > 0 {
		gptMessage = gptResponse.Choices[0].Message.Content
	} else {
		gptMessage = "GPTレスポンスが空です。"
	}

	// SlackBot（GPT）の応答を返す
	err = u.slack.CreateNewBotMessage(channelId, timeStamp, gptMessage)
	if err != nil {
		return fmt.Errorf("failed u.slack.CreateNewBotMessage for channel %s, timestamp %s: %v", channelId, timeStamp, err)
	}

	// 使用量を加算
	currentData.AddTokenUsage(gptResponse.Usage.TotalTokens)
	err = u.ss.UpdateSpreadsheet(ctx, *currentData)
	if err != nil {
		return fmt.Errorf("failed u.ss.UpdateSpreadsheet: %w", err)
	}

	return nil
}
