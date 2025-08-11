package slack

import (
	"fmt"
	"regexp"
	"time"
)

type AnalyzeRequest struct {
	TraceID      string `json:"trace_id"`
	Channel      string `json:"channel"`
	ThreadTS     string `json:"thread_ts"`
	From         string `json:"from,omitempty"`          // Slack workflow format: "August 10th, 2025 at 6:00 PM UTC"
	To           string `json:"to,omitempty"`            // Slack workflow format: "August 10th, 2025 at 6:00 PM UTC"
	AnalysisType string `json:"analysis_type,omitempty"` // "duration" or "error"
}

type AnalyzeResponse struct {
	TraceID string `json:"trace_id"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Config struct {
	// Slack settings
	SlackBotToken      string `envconfig:"SLACK_BOT_TOKEN" required:"true"`
	SlackSigningSecret string `envconfig:"SLACK_SIGNING_SECRET" required:"true"`

	// Cloud Tasks settings
	GoogleCloudProject string `envconfig:"GOOGLE_CLOUD_PROJECT" required:"true"`
	TasksQueueName     string `envconfig:"TASKS_QUEUE_NAME" default:"analyze-queue"`
	WorkerEndpoint     string `envconfig:"WORKER_ENDPOINT" required:"true"`
	TasksLocation      string `envconfig:"TASKS_LOCATION" default:"us-central1"`

	// Backend settings
	TraceBackend string `envconfig:"TRACE_BACKEND" default:"newrelic"`
	LogBackend   string `envconfig:"LOG_BACKEND"`
	LLMBackend   string `envconfig:"LLM_BACKEND" default:"vertexai"`

	// NewRelic settings
	NewRelicAPIKey    string `envconfig:"NEWRELIC_API_KEY"`
	NewRelicAccountID string `envconfig:"NEWRELIC_ACCOUNT_ID"`

	// GCP settings
	GCPProjectID string `envconfig:"GCP_PROJECT_ID"`

	// VertexAI settings
	VertexAIProjectID string `envconfig:"VERTEXAI_PROJECT_ID"`
	VertexAILocation  string `envconfig:"VERTEXAI_LOCATION" default:"us-central1"`

	// Analysis settings
	AnalysisLanguage string `envconfig:"ANALYSIS_LANGUAGE" default:"en"`
}

// ParseSlackWorkflowDate parses Slack workflow date format like "August 10th, 2025 at 6:00 PM UTC"
func ParseSlackWorkflowDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	// Remove ordinal suffixes (1st, 2nd, 3rd, 4th, etc.)
	re := regexp.MustCompile(`(\d)(st|nd|rd|th)`)
	cleanedDate := re.ReplaceAllString(dateStr, "$1")

	// Parse using Go's reference time format
	layout := "January 2, 2006 at 3:04 PM MST"
	
	t, err := time.Parse(layout, cleanedDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse Slack date format '%s': %w", dateStr, err)
	}

	return t, nil
}
