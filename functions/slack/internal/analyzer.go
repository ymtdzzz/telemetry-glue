package slack

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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

// buildCombinedData creates CombinedData from raw trace and log data
func (a *Analyzer) buildCombinedData(traceID string, fromTime, toTime time.Time) (*analyzer.CombinedData, error) {
	aggregator := analyzer.NewDataAggregator()

	// Get and process trace data
	spansJSON, err := a.getTraceData(traceID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get trace data: %w", err)
	}

	if spansJSON != "" {
		if err := a.addSpansToAggregator(aggregator, spansJSON); err != nil {
			return nil, fmt.Errorf("failed to add spans to aggregator: %w", err)
		}
	}

	// Get and process log data
	logsJSON, err := a.getLogData(traceID, fromTime, toTime)
	if err != nil {
		log.Printf("Warning: failed to get log data: %v", err)
		// Continue analysis even if logs cannot be retrieved
	} else if logsJSON != "" {
		if err := a.addLogsToAggregator(aggregator, logsJSON); err != nil {
			log.Printf("Warning: failed to add logs to aggregator: %v", err)
			// Continue analysis even if logs cannot be processed
		}
	}

	return aggregator.GetCombinedData(), nil
}

// addSpansToAggregator adds spans from JSON to the data aggregator
func (a *Analyzer) addSpansToAggregator(aggregator *analyzer.DataAggregator, spansJSON string) error {
	// Parse the spans result format from NewRelic
	var spansResult map[string]interface{}
	if err := json.Unmarshal([]byte(spansJSON), &spansResult); err != nil {
		return fmt.Errorf("failed to parse spans JSON: %w", err)
	}

	// Create a JSON object with spans field to match expected format
	dataObj := map[string]interface{}{
		"spans": spansResult["spans"],
	}

	// Use the aggregator's addJSONObject method
	reader := strings.NewReader(string(mustMarshal(dataObj)))
	return aggregator.ReadFromStdin(reader)
}

// addLogsToAggregator adds logs from JSON to the data aggregator
func (a *Analyzer) addLogsToAggregator(aggregator *analyzer.DataAggregator, logsJSON string) error {
	// Parse the logs result format from GCP
	var logsResult map[string]interface{}
	if err := json.Unmarshal([]byte(logsJSON), &logsResult); err != nil {
		return fmt.Errorf("failed to parse logs JSON: %w", err)
	}

	// Create a JSON object with logs field to match expected format
	dataObj := map[string]interface{}{
		"logs": logsResult["logs"],
	}

	// Use the aggregator's addJSONObject method
	reader := strings.NewReader(string(mustMarshal(dataObj)))
	return aggregator.ReadFromStdin(reader)
}

// mustMarshal marshals to JSON or panics
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return data
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

	// 1. Build combined data from traces and logs
	combinedData, err := a.buildCombinedData(req.TraceID, fromTime, toTime)
	if err != nil {
		return nil, fmt.Errorf("failed to build combined data: %w", err)
	}

	// 2. Determine analysis type
	analysisType := analyzer.AnalysisTypeDuration // default
	if req.AnalysisType == "error" {
		analysisType = analyzer.AnalysisTypeError
	}

	// 3. Perform LLM analysis using unified analyzer
	analysisResult, err := a.performAnalysis(ctx, analysisType, combinedData)
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
		Result:  analysisResult.Content,
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
		spansResult, err := client.SearchSpans(newrelic.SpansRequest{
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
	// Skip log collection if no backend is configured
	if a.config.LogBackend == "" {
		log.Printf("Log backend not configured, skipping log collection")
		return "", nil
	}

	switch a.config.LogBackend {
	case "gcp":
		client, err := gcp.NewClient()
		if err != nil {
			return "", fmt.Errorf("failed to create GCP client: %w", err)
		}

		// Get logs
		logsResult, err := client.SearchLogs(gcp.LogsRequest{
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
		log.Printf("Unsupported log backend '%s', skipping log collection", a.config.LogBackend)
		return "", nil
	}
}

func (a *Analyzer) performAnalysis(ctx context.Context, analysisType analyzer.AnalysisType, data *analyzer.CombinedData) (*analyzer.AnalysisResult, error) {
	switch a.config.LLMBackend {
	case "vertexai":
		// Create VertexAI provider
		provider, err := analyzer.NewVertexAIProvider(
			a.config.VertexAIProjectID,
			a.config.VertexAILocation,
			"gemini-1.5-flash",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create VertexAI provider: %w", err)
		}
		defer func() {
			if err := provider.Close(); err != nil {
				log.Printf("Warning: Failed to close provider: %v", err)
			}
		}()

		// Create unified analyzer using pkg/analyzer
		unifiedAnalyzer := analyzer.NewAnalyzer(provider, "vertexai", "gemini-1.5-flash")

		// Perform analysis with language support
		result, err := unifiedAnalyzer.AnalyzeWithLanguage(ctx, analysisType, data, a.config.AnalysisLanguage)
		if err != nil {
			return nil, fmt.Errorf("failed to perform analysis: %w", err)
		}

		return result, nil

	default:
		return nil, fmt.Errorf("unsupported LLM backend: %s", a.config.LLMBackend)
	}
}
func (a *Analyzer) postResultToSlack(req *AnalyzeRequest, result *analyzer.AnalysisResult) error {
	// Create enhanced message with analysis metadata
	analysisTypeDisplay := "Performance"
	if result.AnalysisType == "error" {
		analysisTypeDisplay = "Error"
	}

	message := fmt.Sprintf(`✅ **%s Analysis Complete**

**Trace ID:** %s
**Data Summary:** %s
**Provider:** %s (%s)

**Analysis Result:**
`+"```"+`
%s
`+"```",
		analysisTypeDisplay,
		req.TraceID,
		result.Summary,
		result.Provider,
		result.Model,
		result.Content)

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
