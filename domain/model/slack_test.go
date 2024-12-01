package model

import (
	"testing"
)

func TestOptimizeMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   SlackMessage
		botUserID string
		want      string
	}{
		{
			name: "Message with bot user ID",
			message: SlackMessage{
				Text: "Hello <@botUserID>!",
				User: "U12345",
			},
			botUserID: "botUserID",
			want:      "Hello <@[GptBot]>!",
		},
		{
			name: "Message without bot user ID",
			message: SlackMessage{
				Text: "Hello world!",
				User: "U12345",
			},
			botUserID: "botUserID",
			want:      "Hello world!",
		},
		{
			name: "Message with leading and trailing spaces",
			message: SlackMessage{
				Text: "  Hello <@botUserID>!  ",
				User: "U12345",
			},
			botUserID: "botUserID",
			want:      "Hello <@[GptBot]>!",
		},
		{
			name: "Message with newlines",
			message: SlackMessage{
				Text: "Hello\n<@botUserID>!",
				User: "U12345",
			},
			botUserID: "botUserID",
			want:      "Hello <@[GptBot]>!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.message.OptimizeMessage(tt.botUserID)
			if got != tt.want {
				t.Errorf("OptimizeMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractConversationFlow(t *testing.T) {
	tests := []struct {
		name      string
		messages  SlackMessages
		botUserID string
		want      []map[string]string
	}{
		{
			name: "Single message",
			messages: SlackMessages{
				{
					Text: "Hello <@botUserID>!",
					User: "U12345",
				},
			},
			botUserID: "botUserID",
			want: []map[string]string{
				{
					"speaker": "U12345",
					"message": "Hello <@[GptBot]>!",
				},
			},
		},
		{
			name: "Multiple messages",
			messages: SlackMessages{
				{
					Text: "Hello <@botUserID>!",
					User: "U12345",
				},
				{
					Text: "How are you?",
					User: "U67890",
				},
			},
			botUserID: "botUserID",
			want: []map[string]string{
				{
					"speaker": "U12345",
					"message": "Hello <@[GptBot]>!",
				},
				{
					"speaker": "U67890",
					"message": "How are you?",
				},
			},
		},
		{
			name:      "No messages",
			messages:  SlackMessages{},
			botUserID: "botUserID",
			want:      []map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.messages.extractConversationFlow(tt.botUserID)
			if len(got) != len(tt.want) {
				t.Errorf("extractConversationFlow() length = %v, want %v", len(got), len(tt.want))
			}
			for i := range got {
				if got[i]["speaker"] != tt.want[i]["speaker"] || got[i]["message"] != tt.want[i]["message"] {
					t.Errorf("extractConversationFlow() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
