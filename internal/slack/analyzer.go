package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"
	"github.com/ymtdzzz/telemetry-glue/pkg/analyzer"
	"github.com/ymtdzzz/telemetry-glue/pkg/backend/gcp"
	"github.com/ymtdzzz/telemetry-glue/pkg/backend/newrelic"
)

type Analyzer struct {
	config      *Config
	slackClient *slack.Client
}

func NewAnalyzer(config *Config) *Analyzer {
	slackClient := slack.New(config.SlackBotToken)

	return &Analyzer{
		config:      config,
		slackClient: slackClient,
	}
}

func (a *Analyzer) AnalyzeTrace(ctx context.Context, req *AnalyzeRequest) (*AnalyzeResponse, error) {
	log.Printf("Starting analysis for trace_id: %s", req.TraceID)

	// Parse from/to dates if provided
	var fromTime, toTime time.Time
	var err error
	
	if req.From != "" {
		fromTime, err = ParseSlackWorkflowDate(req.From)
		if err != nil {
			return nil, fmt.Errorf("failed to parse 'from' date: %w", err)
		}
	} else {
		// Default to 1 hour ago
		fromTime = time.Now().Add(-1 * time.Hour)
	}
	
	if req.To != "" {
		toTime, err = ParseSlackWorkflowDate(req.To)
		if err != nil {
			return nil, fmt.Errorf("failed to parse 'to' date: %w", err)
		}
	} else {
		// Default to now
		toTime = time.Now()
	}

	log.Printf("Time range: %s to %s", fromTime.Format(time.RFC3339), toTime.Format(time.RFC3339))

	// 1. Get trace data
	spans, err := a.getTraceData(req.TraceID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace data: %w", err)
	}

	// 2. Get log data
	logs, err := a.getLogData(req.TraceID, fromTime, toTime)
	if err != nil {
		log.Printf("Warning: failed to get log data: %v", err)
		// Continue analysis even if logs cannot be retrieved
		logs = "[]"
	}

	// 3. Perform LLM analysis
	analysisResult, err := a.performAnalysis(ctx, spans, logs)
	if err != nil {
		return nil, fmt.Errorf("failed to perform analysis: %w", err)
	}

	// 4. Post results to Slack
	if err := a.postResultToSlack(req, analysisResult); err != nil {
		log.Printf("Failed to post result to Slack: %v", err)
		// Slack posting failure doesn't affect analysis results
	}

	return &AnalyzeResponse{
		TraceID: req.TraceID,
		Status:  "completed",
		Result:  analysisResult,
	}, nil
}

func (a *Analyzer) getTraceData(traceID string, fromTime, toTime time.Time) (string, error) {
	switch a.config.TraceBackend {
	case "newrelic":
		// Environment variables must be set beforehand
		client, err := newrelic.NewClient()
		if err != nil {
			return "", fmt.Errorf("failed to create NewRelic client: %w", err)
		}

		// Use provided time range
		timeRange := newrelic.TimeRange{
			Start: fromTime,
			End:   toTime,
		}

		// Get spans
		spansResult, _, err := client.SearchSpans(newrelic.SpansRequest{
			TraceID:   traceID,
			TimeRange: timeRange,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get spans from NewRelic: %w", err)
		}

		// Serialize to JSON format
		result, err := json.Marshal(spansResult)
		if err != nil {
			return "", fmt.Errorf("failed to marshal spans result: %w", err)
		}
		return string(result), nil

	default:
		return "", fmt.Errorf("unsupported trace backend: %s", a.config.TraceBackend)
	}
}

func (a *Analyzer) getLogData(traceID string, fromTime, toTime time.Time) (string, error) {
	switch a.config.LogBackend {
	case "gcp":
		client, err := gcp.NewClient()
		if err != nil {
			return "", fmt.Errorf("failed to create GCP client: %w", err)
		}

		// Get logs
		logsResult, _, err := client.SearchLogs(gcp.LogsRequest{
			ProjectID: a.config.GCPProjectID,
			TraceID:   traceID,
			Limit:     100,
			StartTime: fromTime,
			EndTime:   toTime,
		})
		if err != nil {
			return "", fmt.Errorf("failed to get logs from GCP: %w", err)
		}

		// Serialize to JSON format
		result, err := json.Marshal(logsResult)
		if err != nil {
			return "", fmt.Errorf("failed to marshal logs result: %w", err)
		}
		return string(result), nil

	default:
		return "", fmt.Errorf("unsupported log backend: %s", a.config.LogBackend)
	}
}

func (a *Analyzer) performAnalysis(ctx context.Context, spans, logs string) (string, error) {
	switch a.config.LLMBackend {
	case "vertexai":
		// Create VertexAI provider
		provider, err := analyzer.NewVertexAIProvider(
			a.config.VertexAIProjectID,
			a.config.VertexAILocation,
			"gemini-1.5-flash", // Default model
		)
		if err != nil {
			return "", fmt.Errorf("failed to create VertexAI provider: %w", err)
		}
		defer func() {
			if err := provider.Close(); err != nil {
				log.Printf("Warning: Failed to close provider: %v", err)
			}
		}()

		// Prepare analysis prompt
		prompt := fmt.Sprintf(`
Please analyze the following telemetry data and provide insights about potential issues, performance bottlenecks, or anomalies.

Spans Data:
%s

Logs Data:
%s

Please provide:
1. Summary of the trace flow
2. Any detected issues or errors
3. Performance analysis
4. Recommendations for improvement
`, spans, logs)

		result, err := provider.GenerateContent(ctx, prompt)
		if err != nil {
			return "", fmt.Errorf("failed to perform analysis: %w", err)
		}

		return result, nil

	default:
		return "", fmt.Errorf("unsupported LLM backend: %s", a.config.LLMBackend)
	}
}
func (a *Analyzer) postResultToSlack(req *AnalyzeRequest, result string) error {
	message := fmt.Sprintf("✅ **Trace Analysis Complete**\n\n**Trace ID:** `%s`\n\n**Analysis Result:**\n```\n%s\n```", req.TraceID, result)

	_, _, err := a.slackClient.PostMessage(
		req.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(req.ThreadTS),
	)

	if err != nil {
		return fmt.Errorf("failed to post message to Slack: %w", err)
	}

	return nil
}

func (a *Analyzer) PostErrorToSlack(req *AnalyzeRequest, errorMsg string) error {
	message := fmt.Sprintf("❌ **Trace Analysis Failed**\n\n**Trace ID:** `%s`\n\n**Error:**\n```\n%s\n```", req.TraceID, errorMsg)

	_, _, err := a.slackClient.PostMessage(
		req.Channel,
		slack.MsgOptionText(message, false),
		slack.MsgOptionTS(req.ThreadTS),
	)

	return err
}
