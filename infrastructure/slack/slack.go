package slack

import (
	"fmt"

	"github.com/gs1068/slack-gpt-bot/domain/repository"
	"github.com/slack-go/slack"
)

var SlackBotUserID string

type slackRepository struct {
	slackClient *slack.Client
}

func NewSlackRepository(slackClient *slack.Client) repository.SlackRepository {
	return &slackRepository{
		slackClient: slackClient,
	}
}

func SlackClient(slackBotToken string) *slack.Client {
	client := slack.New(slackBotToken)
	return client
}

func (r *slackRepository) LoadConversationReplies(channelId string, timeStamp string) ([]slack.Message, error) {
	var messages []slack.Message

	var cursor string = ""
	for {
		resp, hasMore, nextCursor, err := r.slackClient.GetConversationReplies(&slack.GetConversationRepliesParameters{
			ChannelID: channelId,
			Timestamp: timeStamp,
			Cursor:    cursor,
		})

		if err != nil {
			return []slack.Message{}, fmt.Errorf("failed to get conversation history: %v", err)
		}

		messages = append(messages, resp...)

		if !hasMore {
			break
		}

		cursor = nextCursor
	}

	return messages, nil
}

func (r *slackRepository) GetBotUserId() (string, error) {
	authTestResponse, err := r.slackClient.AuthTest()
	if err != nil {
		return "", fmt.Errorf("failed r.slackClient.AuthTest: %w", err)
	}

	return authTestResponse.UserID, nil
}

func (r *slackRepository) CreateNewBotMessage(channelId string, timeStamp string, msg string) error {
	_, _, err := r.slackClient.PostMessage(
		channelId,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionTS(timeStamp),
	)
	if err != nil {
		return fmt.Errorf("failed r.slackClient.PostMessage: %w", err)
	}

	return nil
}
