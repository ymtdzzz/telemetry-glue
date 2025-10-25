package logger

import "github.com/slack-go/slack"

// SlackLogger is a logger that sends logs to a Slack channel
type SlackLogger struct {
	client    *slack.Client
	channelID string
	threadTS  string
}

func NewSlackLogger(client *slack.Client, channelID, threadTS string) *SlackLogger {
	return &SlackLogger{
		client:    client,
		channelID: channelID,
		threadTS:  threadTS,
	}
}

func (l *SlackLogger) Log(message string) error {
	_, _, err := l.client.PostMessage(
		l.channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(l.threadTS),
	)
	return err
}
