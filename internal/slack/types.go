package slack

type AnalyzeRequest struct {
	TraceID  string `json:"trace_id"`
	Channel  string `json:"channel"`
	ThreadTS string `json:"thread_ts"`
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
	LogBackend   string `envconfig:"LOG_BACKEND" default:"gcp"`
	LLMBackend   string `envconfig:"LLM_BACKEND" default:"vertexai"`

	// NewRelic settings
	NewRelicAPIKey    string `envconfig:"NEWRELIC_API_KEY"`
	NewRelicAccountID string `envconfig:"NEWRELIC_ACCOUNT_ID"`

	// GCP settings
	GCPProjectID string `envconfig:"GCP_PROJECT_ID"`

	// VertexAI settings
	VertexAIProjectID string `envconfig:"VERTEXAI_PROJECT_ID"`
	VertexAILocation  string `envconfig:"VERTEXAI_LOCATION" default:"us-central1"`
}
