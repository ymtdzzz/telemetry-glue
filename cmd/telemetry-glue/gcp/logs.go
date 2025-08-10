package gcp

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/gcp"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
	"github.com/ymtdzzz/telemetry-glue/internal/pipeline"
)

// LogsFlags holds GCP-specific flags for logs command
type LogsFlags struct {
	Common    common.CommonFlags
	ProjectID string
	TraceID   string
	Limit     int
}

// LogsCmd creates the logs subcommand for GCP
func LogsCmd() *cobra.Command {
	flags := &LogsFlags{}

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get log entries for a specific trace ID in GCP Cloud Logging",
		Long: `Get log entries for a specific trace ID in GCP Cloud Logging.
This command retrieves all log entries associated with a specific trace ID
ordered by timestamp (newest first).

Examples:
  # Get logs for a trace
  telemetry-glue gcp logs --project-id my-project --trace-id abc123def456
  
  # Get logs with limit
  telemetry-glue gcp logs --project-id my-project --trace-id abc123def456 --limit 50
  
  # Get logs with time range
  telemetry-glue gcp logs --project-id my-project --trace-id abc123def456 --time-range "2024-01-15T10:00:00Z,2024-01-15T11:00:00Z"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(flags)
		},
	}

	// Add GCP-specific flags
	cmd.Flags().StringVarP(&flags.ProjectID, "project-id", "p", "", "GCP Project ID (required)")
	cmd.Flags().StringVar(&flags.TraceID, "trace-id", "", "Trace ID to get logs from (required)")
	cmd.Flags().IntVarP(&flags.Limit, "limit", "l", 100, "Maximum number of log entries to return")

	// Add common flags
	common.AddCommonFlags(cmd, &flags.Common)

	// Mark required flags
	cmd.MarkFlagRequired("project-id")
	cmd.MarkFlagRequired("trace-id")

	return cmd
}

func runLogs(flags *LogsFlags) error {
	// Create passthrough handler for pipeline support
	passthroughHandler := pipeline.NewPassthroughHandler()

	// Read any existing data from stdin
	existingData, err := passthroughHandler.ReadStdinIfAvailable()
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	// Parse time range
	timeRange, err := common.ParseTimeRange(flags.Common.TimeRange)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	// Parse output format
	format, err := common.ParseFormat(flags.Common.Format)
	if err != nil {
		return err
	}

	// Create GCP client
	client, err := gcp.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create GCP client: %w", err)
	}

	// Execute logs search
	entries, webLink, err := client.SearchLogs(gcp.LogsRequest{
		ProjectID: flags.ProjectID,
		TraceID:   flags.TraceID,
		Limit:     flags.Limit,
		StartTime: timeRange.Start,
		EndTime:   timeRange.End,
	})
	if err != nil {
		return fmt.Errorf("failed to search logs: %w", err)
	}

	// Convert GCP LogEntry to output LogEntry
	var outputLogs []output.LogEntry
	for _, entry := range entries {
		logEntry := output.LogEntry{
			Timestamp:  entry.Timestamp,
			TraceID:    extractTraceID(entry.Trace),
			SpanID:     entry.SpanID,
			Attributes: make(map[string]interface{}),
		}

		// Determine message from payload
		if entry.TextPayload != "" {
			logEntry.Message = entry.TextPayload
		} else if entry.JSONPayload != nil {
			// Try to extract message from JSON payload
			if msg, ok := entry.JSONPayload["message"].(string); ok {
				logEntry.Message = msg
			} else if msg, ok := entry.JSONPayload["msg"].(string); ok {
				logEntry.Message = msg
			} else {
				// If no message field, use the entire JSON as string
				logEntry.Message = fmt.Sprintf("JSON: %v", entry.JSONPayload)
			}
			// Add JSON payload to attributes
			for k, v := range entry.JSONPayload {
				if k != "message" && k != "msg" {
					logEntry.Attributes[k] = v
				}
			}
		} else if entry.ProtoPayload != nil {
			logEntry.Message = fmt.Sprintf("Proto: %v", entry.ProtoPayload)
			// Add proto payload to attributes
			for k, v := range entry.ProtoPayload {
				logEntry.Attributes[k] = v
			}
		}

		// Add other fields to attributes
		if entry.Severity != "" {
			logEntry.Attributes["severity"] = entry.Severity
		}
		if entry.LogName != "" {
			logEntry.Attributes["log_name"] = entry.LogName
		}
		if entry.Resource != nil {
			logEntry.Attributes["resource"] = entry.Resource
		}
		if entry.Labels != nil {
			for k, v := range entry.Labels {
				logEntry.Attributes["label_"+k] = v
			}
		}

		outputLogs = append(outputLogs, logEntry)
	}

	// Create logs result
	result := output.LogsResult{
		Logs:    outputLogs,
		WebLink: webLink,
	}

	// Merge with existing data and output
	mergedData := passthroughHandler.MergeLogsResult(existingData, &result)
	return passthroughHandler.OutputMergedResult(mergedData, &result, format)
}

// extractTraceID extracts the trace ID from the full trace resource path
// e.g., "projects/my-project/traces/abc123" -> "abc123"
func extractTraceID(traceResource string) string {
	if traceResource == "" {
		return ""
	}
	parts := strings.Split(traceResource, "/")
	if len(parts) >= 4 && parts[2] == "traces" {
		return parts[3]
	}
	return traceResource
}
