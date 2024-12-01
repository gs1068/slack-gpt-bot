package model

import (
	"fmt"
	"strings"

	"github.com/slack-go/slack"
)

const (
	LimitMessage     = "本日の利用制限を超えました。明日以降に再度お試しください。"
	MaxFetchMessages = 20
)

type SlackMessage struct {
	Text string // メッセージの内容
	User string // メッセージを送信したユーザーのID
}

type SlackMessages []SlackMessage

type BotMessage struct {
	Client       *slack.Client
	ChannelID    string
	OutputTS     string
	ControllerTS string
}

func (m *SlackMessage) OptimizeMessage(botUserID string) string {
	// botの場合はbotであることもわかるようにIDとBOTをセットでメッセージを整形
	text := strings.ReplaceAll(m.Text, botUserID, "[GptBot]")
	text = strings.TrimSpace(text)
	text = formatMessage(text)
	return text
}

func formatMessage(text string) string {
	return strings.ReplaceAll(text, "\n", " ") // 改行をスペースに置換
}

func (messages SlackMessages) extractConversationFlow(botUserID string) []map[string]string {
	var flow []map[string]string
	for _, message := range messages {
		flow = append(flow, map[string]string{
			"speaker": message.User,                       // 発言者を取得
			"message": message.OptimizeMessage(botUserID), // メッセージを整形
		})
	}
	return flow
}

func (messages SlackMessages) LimitMessages(maxMessages int) SlackMessages {
	if len(messages) > maxMessages {
		return messages[len(messages)-maxMessages:] // 最新の`maxMessages`件のみ保持
	}
	return messages
}

func (messages SlackMessages) CreatePrompt(botUserID string) string {
	messages = messages.LimitMessages(MaxFetchMessages)

	var builder strings.Builder
	builder.WriteString("以下はSlackスレッドの履歴を含んだGPTプロンプトです。下記を踏まえて答えてください\n")
	for _, flow := range messages.extractConversationFlow(botUserID) {
		builder.WriteString(fmt.Sprintf("%s, message: %s\n", flow["speaker"], flow["message"]))
	}
	return builder.String()
}

func ConvertToSlackMessages(messages []slack.Message) SlackMessages {
	var slackMessages SlackMessages
	for _, message := range messages {
		slackMessages = append(slackMessages, SlackMessage{
			Text: message.Text,
			User: message.User,
		})
	}
	return slackMessages
}
