package gcp

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/pkg/backend/gcp"
	"github.com/ymtdzzz/telemetry-glue/pkg/pipeline"
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
	if err := cmd.MarkFlagRequired("project-id"); err != nil {
		panic(fmt.Sprintf("Failed to mark project-id flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("trace-id"); err != nil {
		panic(fmt.Sprintf("Failed to mark trace-id flag as required: %v", err))
	}

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
	result, err := client.SearchLogs(gcp.LogsRequest{
		ProjectID: flags.ProjectID,
		TraceID:   flags.TraceID,
		Limit:     flags.Limit,
		StartTime: timeRange.Start,
		EndTime:   timeRange.End,
	})
	if err != nil {
		return fmt.Errorf("failed to search logs: %w", err)
	}

	// Merge with existing data and output
	mergedData := passthroughHandler.MergeLogsResult(existingData, result)
	return passthroughHandler.OutputMergedResult(mergedData, result, format)
}
