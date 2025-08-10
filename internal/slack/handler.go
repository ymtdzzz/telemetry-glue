package slack

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

type Handler struct {
	slack       *slack.Client
	socketMode  *socketmode.Client
	tasksClient *TasksClient
}

func NewHandler(config *Config, tasksClient *TasksClient) *Handler {
	api := slack.New(config.SlackBotToken, slack.OptionDebug(true))
	socketClient := socketmode.New(api, socketmode.OptionDebug(true))

	return &Handler{
		slack:       api,
		socketMode:  socketClient,
		tasksClient: tasksClient,
	}
}

func (h *Handler) Start(ctx context.Context) error {
	go func() {
		for evt := range h.socketMode.Events {
			switch evt.Type {
			case socketmode.EventTypeSlashCommand:
				cmd, ok := evt.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("Ignored %+v\n", evt)
					continue
				}
				h.handleSlashCommand(ctx, &cmd, evt.Request)

			case socketmode.EventTypeEventsAPI:
				eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Ignored %+v\n", evt)
					continue
				}
				h.socketMode.Ack(*evt.Request)
				h.handleEventMessage(eventsAPIEvent)
			}
		}
	}()

	return h.socketMode.Run()
}

func (h *Handler) handleSlashCommand(ctx context.Context, cmd *slack.SlashCommand, req *socketmode.Request) {
	h.socketMode.Ack(*req)

	switch cmd.Command {
	case "/analyze-trace":
		h.handleAnalyzeTrace(ctx, cmd)
	default:
		log.Printf("Unknown command: %s", cmd.Command)
	}
}

func (h *Handler) handleAnalyzeTrace(ctx context.Context, cmd *slack.SlashCommand) {
	traceID := strings.TrimSpace(cmd.Text)
	if traceID == "" {
		h.postMessage(cmd.ChannelID, "❌ Please specify trace_id. Example: `/analyze-trace abc123def456`", "")
		return
	}

	// Basic validation for trace_id format (alphanumeric characters)
	if !isValidTraceID(traceID) {
		h.postMessage(cmd.ChannelID, "❌ Invalid trace_id format.", "")
		return
	}

	// Post analysis start message
	timestamp, err := h.postMessage(cmd.ChannelID, "🔍 Starting analysis...", "")
	if err != nil {
		log.Printf("Failed to post start message: %v", err)
		return
	}

	// Enqueue job to Cloud Tasks
	analyzeReq := &AnalyzeRequest{
		TraceID:  traceID,
		Channel:  cmd.ChannelID,
		ThreadTS: timestamp,
	}

	if err := h.tasksClient.EnqueueAnalyzeTask(ctx, analyzeReq); err != nil {
		log.Printf("Failed to enqueue task: %v", err)
		h.postMessage(cmd.ChannelID, "❌ Failed to enqueue analysis task.", timestamp)
		return
	}

	log.Printf("Analysis task enqueued for trace_id: %s, channel: %s", traceID, cmd.ChannelID)
}

func (h *Handler) handleEventMessage(event slackevents.EventsAPIEvent) {
	switch event.Type {
	case slackevents.CallbackEvent:
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.MessageEvent:
			// Implement message event handling if needed
			log.Printf("Message event: %+v", ev)
		}
	default:
		log.Printf("Unsupported event type: %s", event.Type)
	}
}

func (h *Handler) postMessage(channel, text, threadTS string) (string, error) {
	options := []slack.MsgOption{
		slack.MsgOptionText(text, false),
	}

	if threadTS != "" {
		options = append(options, slack.MsgOptionTS(threadTS))
	}

	_, timestamp, err := h.slack.PostMessage(channel, options...)
	if err != nil {
		return "", fmt.Errorf("failed to post message: %w", err)
	}

	return timestamp, nil
}

func isValidTraceID(traceID string) bool {
	// Basic alphanumeric and hyphen validation (add stricter validation as needed)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, traceID)
	return matched && len(traceID) > 0
}
