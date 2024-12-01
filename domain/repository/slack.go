package repository

import (
	"github.com/slack-go/slack"
)

type SlackRepository interface {
	LoadConversationReplies(channelId string, timeStamp string) ([]slack.Message, error)
	CreateNewBotMessage(channelId string, timeStamp string, msg string) error
	GetBotUserId() (string, error)
}
