package handler

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub/v2"
	"github.com/slack-go/slack"
	"github.com/ymtdzzz/telemetry-glue/pkg/app"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/logger"
	"github.com/ymtdzzz/telemetry-glue/pkg/glue/backend"
)

// Slack bot needs chat:write scope and be added to the channel where it will post messages.

func HandleCommand(w http.ResponseWriter, r *http.Request) {
	slackbotToken := os.Getenv("SLACK_BOT_TOKEN")
	verificationToken := os.Getenv("SLACK_VERIFICATION_TOKEN")
	projectID := os.Getenv("GCP_PROJECT_ID")
	topicID := os.Getenv("GCP_PUBSUB_TOPIC_ID")

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Println("Failed to parse slash command:", err)
		http.Error(w, "Failed to parse slash command", http.StatusInternalServerError)
		return
	}

	if !s.ValidateToken(verificationToken) {
		log.Println("Invalid verification token")
		http.Error(w, "Invalid verification token", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Println("Failed to create Pub/Sub client:", err)
		http.Error(w, "Failed to create Pub/Sub client", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	switch s.Command {
	case "/telemetry-glue":
		slackClient := slack.New(slackbotToken)
		_, ts, err := slackClient.PostMessage(
			s.ChannelID,
			slack.MsgOptionText("Processing your request...", false),
		)
		if err != nil {
			log.Println("Failed to post initial message to Slack:", err)
			http.Error(w, "Failed to post initial message to Slack", http.StatusInternalServerError)
			return
		}

		// parse /telemetry-glue analyze <trace-id> <date yyyy/mm/dd> <time HH:MM>
		// example: /telemetry-glue analyze 1234567890abcdef 2024/05/12 15:10
		args := strings.Split(s.Text, " ")
		if len(args) == 1 {
			if args[0] == "help" {
				helpMsg := "使い方: /telemetry-glue analyze <trace-id> <date yyyy/mm/dd> <time HH:MM>\n" +
					"例: /telemetry-glue analyze 1234567890abcdef 2024/05/12 15:10"
				_, _, err := slackClient.PostMessage(
					s.ChannelID,
					slack.MsgOptionText(helpMsg, false),
					slack.MsgOptionTS(ts),
				)
				if err != nil {
					log.Println("Failed to post help message to Slack:", err)
					http.Error(w, "Failed to post help message to Slack", http.StatusInternalServerError)
				}
			}
			return
		}
		if len(args) != 4 || args[0] != "analyze" {
			log.Printf("Invalid command format: %s", s.Text)
			http.Error(w, "使い方が違うみたい。/telemetry-glue helpを確認してね", http.StatusBadRequest)
			return
		}
		traceID := args[1]

		log.Printf("channel_id: %s, thread_ts: %s, trace_id: %s", s.ChannelID, ts, traceID)

		publisher := client.Publisher(topicID)
		result := publisher.Publish(ctx, &pubsub.Message{
			Data: []byte(s.Text),
			Attributes: map[string]string{
				"channel_id": s.ChannelID,
				"thread_ts":  ts,
				"trace_id":   traceID,
				"timestamp":  args[2] + " " + args[3],
			},
		})
		_, err = result.Get(ctx)
		if err != nil {
			log.Println("Failed to publish message to Pub/Sub:", err)
			http.Error(w, "Failed to publish message to Pub/Sub", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
		return
	}
}

func HandlePubsub(ctx context.Context, m *pubsub.Message) error {
	slackbotToken := os.Getenv("SLACK_BOT_TOKEN")
	if slackbotToken == "" {
		return errors.New("SLACK_BOT_TOKEN is not set")
	}

	channelID := m.Attributes["channel_id"]
	threadTS := m.Attributes["thread_ts"]
	traceID := m.Attributes["trace_id"]
	timestamp := m.Attributes["timestamp"]
	if channelID == "" || threadTS == "" || traceID == "" || timestamp == "" {
		return errors.New("missing channel_id or thread_ts or trace_id or timestamp in message attributes")
	}

	client := slack.New(slackbotToken)
	logger := logger.NewSlackLogger(client, channelID, threadTS)

	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("failed to load JST location: %w", err)
	}
	start, err := time.ParseInLocation("2006/01/02 15:04", timestamp, jst)
	if err != nil {
		logger.Log("日付の解析に失敗しました。フォーマットはyyyy/mm/dd HH:MMで指定してください。")
		return fmt.Errorf("failed to parse timestamp: %w", err)
	}

	timeRange := backend.TimeRange{
		Start: start,
		End:   start.Add(30 * time.Minute),
	}

	app, err := app.NewApp("", logger, traceID, &timeRange)
	if err != nil {
		logger.Log("Appの初期化に失敗しました。")
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	return app.RunDuration(context.Background(), false)
}
