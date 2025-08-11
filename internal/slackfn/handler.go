package slackfn

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	slackInternal "github.com/ymtdzzz/telemetry-glue/internal/slack"
)

// HandleSlackEvent handles Slack events (app mentions, etc.)
func HandleSlackEvent(w http.ResponseWriter, r *http.Request) {
	// Accept only POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load configuration from environment variables
	var config slackInternal.Config
	if err := envconfig.Process("", &config); err != nil {
		log.Printf("Failed to load configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Parse Slack event
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Printf("Failed to parse event: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Handle URL verification challenge
	if eventsAPIEvent.Type == slackevents.URLVerification {
		var r slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &r); err != nil {
			log.Printf("Failed to unmarshal challenge: %v", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text")
		if _, err := w.Write([]byte(r.Challenge)); err != nil {
			log.Printf("Failed to write challenge response: %v", err)
		}
		return
	}

	// Handle slash commands and events
	switch eventsAPIEvent.Type {
	case slackevents.CallbackEvent:
		handleSlackCallback(w, r, &eventsAPIEvent, &config)
	default:
		log.Printf("Unsupported event type: %s", eventsAPIEvent.Type)
		w.WriteHeader(http.StatusOK)
	}
}

// HandleAnalyzeTrace handles trace analysis requests
func HandleAnalyzeTrace(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Accept only POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Load configuration from environment variables
	var config slackInternal.Config
	if err := envconfig.Process("", &config); err != nil {
		log.Printf("Failed to load configuration: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Read request body
	var req slackInternal.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create worker analyzer
	analyzer := slackInternal.NewAnalyzer(&config)

	// Execute analysis
	response, err := analyzer.AnalyzeTrace(ctx, &req)
	if err != nil {
		log.Printf("Analysis failed for trace_id %s: %v", req.TraceID, err)

		// Post error to Slack
		errorResponse := &slackInternal.AnalyzeResponse{
			TraceID: req.TraceID,
			Status:  "error",
			Error:   err.Error(),
		}

		// Post error to Slack
		if postErr := analyzer.PostErrorToSlack(&req, err.Error()); postErr != nil {
			log.Printf("Failed to post error to Slack: %v", postErr)
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Analysis completed successfully for trace_id: %s", req.TraceID)
}

func handleSlackCallback(w http.ResponseWriter, r *http.Request, event *slackevents.EventsAPIEvent, config *slackInternal.Config) {
	innerEvent := event.InnerEvent

	switch ev := innerEvent.Data.(type) {
	case *slackevents.MessageEvent:
		// Handle message events if needed
		log.Printf("Message event: %+v", ev)

	case *slackevents.AppMentionEvent:
		// Handle app mention events
		handleAppMention(ev, config)

	default:
		log.Printf("Unsupported inner event type: %T", ev)
	}

	w.WriteHeader(http.StatusOK)
}

func handleAppMention(event *slackevents.AppMentionEvent, config *slackInternal.Config) {
	ctx := context.Background()

	// Parse trace ID from message text
	text := strings.TrimSpace(event.Text)
	traceID := extractTraceID(text)

	if traceID == "" {
		postErrorMessage(config, event.Channel, event.TimeStamp, "❌ Please specify trace_id. Example: `@bot analyze abc123def456`")
		return
	}

	// Basic validation for trace_id format
	if !isValidTraceID(traceID) {
		postErrorMessage(config, event.Channel, event.TimeStamp, "❌ Invalid trace_id format.")
		return
	}

	// Post analysis start message
	slackAPI := slack.New(config.SlackBotToken)
	_, timestamp, err := slackAPI.PostMessage(event.Channel,
		slack.MsgOptionText("🔍 Starting analysis...", false),
		slack.MsgOptionTS(event.TimeStamp))
	if err != nil {
		log.Printf("Failed to post start message: %v", err)
		return
	}

	// Create Cloud Tasks client and enqueue analysis
	tasksClient, err := slackInternal.NewTasksClient(ctx, config)
	if err != nil {
		log.Printf("Failed to create tasks client: %v", err)
		postErrorMessage(config, event.Channel, timestamp, "❌ Failed to create tasks client.")
		return
	}
	defer func() {
		if err := tasksClient.Close(); err != nil {
			log.Printf("Failed to close tasks client: %v", err)
		}
	}()

	// Enqueue analysis task
	analyzeReq := &slackInternal.AnalyzeRequest{
		TraceID:  traceID,
		Channel:  event.Channel,
		ThreadTS: timestamp,
	}

	if err := tasksClient.EnqueueAnalyzeTask(ctx, analyzeReq); err != nil {
		log.Printf("Failed to enqueue task: %v", err)
		postErrorMessage(config, event.Channel, timestamp, "❌ Failed to enqueue analysis task.")
		return
	}

	log.Printf("Analysis task enqueued for trace_id: %s, channel: %s", traceID, event.Channel)
}

func extractTraceID(text string) string {
	// Remove bot mention and extract trace ID
	// Expected format: "@bot analyze <trace_id>" or "@bot <trace_id>"
	words := strings.Fields(text)

	for i, word := range words {
		// Skip the bot mention (first word typically)
		if i == 0 && strings.HasPrefix(word, "<@") {
			continue
		}
		// Skip "analyze" keyword if present
		if strings.ToLower(word) == "analyze" {
			continue
		}
		// Return first potential trace ID
		if len(word) > 0 && isValidTraceID(word) {
			return word
		}
	}

	return ""
}

func postErrorMessage(config *slackInternal.Config, channel, threadTS, message string) {
	slackAPI := slack.New(config.SlackBotToken)
	options := []slack.MsgOption{
		slack.MsgOptionText(message, false),
	}

	if threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}

	_, _, err := slackAPI.PostMessage(channel, options...)
	if err != nil {
		log.Printf("Failed to post error message: %v", err)
	}
}

func isValidTraceID(traceID string) bool {
	// Basic alphanumeric and hyphen validation
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, traceID)
	return matched && len(traceID) > 0
}
