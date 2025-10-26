package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub/v2"
	"github.com/slack-go/slack"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/logger"
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

		log.Printf("channel_id: %s, thread_ts: %s", s.ChannelID, ts)

		publisher := client.Publisher(topicID)
		result := publisher.Publish(ctx, &pubsub.Message{
			Data: []byte(s.Text),
			Attributes: map[string]string{
				"channel_id": s.ChannelID,
				"thread_ts":  ts,
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
	if channelID == "" || threadTS == "" {
		return errors.New("missing channel_id or thread_ts in message attributes")
	}

	client := slack.New(slackbotToken)
	logger := logger.NewSlackLogger(client, channelID, threadTS)
	if err := logger.Log("Received message: " + string(m.Data)); err != nil {
		return err
	}

	return nil
}
