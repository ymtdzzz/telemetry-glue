package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/pubsub/v2"
	"github.com/slack-go/slack"
)

func HandleCommand(w http.ResponseWriter, r *http.Request) {
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
		params := &slack.Msg{ResponseType: "in_channel", Text: "called. args: " + s.Text}
		b, err := json.Marshal(params)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		publisher := client.Publisher(topicID)
		result := publisher.Publish(ctx, &pubsub.Message{
			Data: []byte(s.Text),
		})
		_, err = result.Get(ctx)
		if err != nil {
			log.Println("Failed to publish message to Pub/Sub:", err)
			http.Error(w, "Failed to publish message to Pub/Sub", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	default:
		http.Error(w, "Unknown command", http.StatusBadRequest)
		return
	}
}

func HandlePubsub(ctx context.Context, m *pubsub.Message) error {
	log.Println(string(m.Data))
	if len(m.Attributes) > 0 {
		for k, v := range m.Attributes {
			log.Println("key", k)
			log.Println("value", v)
		}
	}
	return nil
}
