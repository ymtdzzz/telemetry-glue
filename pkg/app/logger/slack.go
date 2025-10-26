package logger

import "github.com/slack-go/slack"

// SlackLogger is a logger that sends logs to a Slack channel
type SlackLogger struct {
	client      *slack.Client
	channelID   string
	responseURL string
}

func NewSlackLogger(client *slack.Client, channelID, responseURL string) *SlackLogger {
	return &SlackLogger{
		client:      client,
		channelID:   channelID,
		responseURL: responseURL,
	}
}

func (l *SlackLogger) Log(message string) error {
	_, _, err := l.client.PostMessage(
		l.channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionResponseURL(l.responseURL, "in_channel"),
	)
	return err
}
