package newrelic

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
	"github.com/ymtdzzz/telemetry-glue/internal/pipeline"
)

// SpansFlags holds NewRelic-specific flags for spans command
type SpansFlags struct {
	Common  common.CommonFlags
	TraceID string
}

// SpansCmd creates the spans subcommand for NewRelic
func SpansCmd() *cobra.Command {
	flags := &SpansFlags{}

	cmd := &cobra.Command{
		Use:   "spans",
		Short: "Get all spans for a specific trace in NewRelic",
		Long: `Get all spans for a specific trace in NewRelic.
This command retrieves all spans within a specific trace ID and displays them
ordered by timestamp (earliest first).

Examples:
  # Get all spans for a trace
  telemetry-glue newrelic spans --trace-id d8a60536187fa0927e45911f8c0dd64b
  
  # Get spans with custom time range
  telemetry-glue newrelic spans --trace-id d8a60536187fa0927e45911f8c0dd64b --time-range "2024-01-15T10:00:00Z,2024-01-15T11:00:00Z"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSpans(flags)
		},
	}

	// Add NewRelic-specific flags
	cmd.Flags().StringVarP(&flags.TraceID, "trace-id", "t", "", "Trace ID to get spans from (required)")

	// Add common flags
	common.AddCommonFlags(cmd, &flags.Common)

	// Mark required flags
	cmd.MarkFlagRequired("trace-id")

	return cmd
}

func runSpans(flags *SpansFlags) error {
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

	// Create NewRelic client
	client, err := newrelic.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create NewRelic client: %w", err)
	}

	// Execute spans search
	spans, webLink, err := client.SearchSpans(newrelic.SpansRequest{
		TraceID: flags.TraceID,
		TimeRange: newrelic.TimeRange{
			Start: timeRange.Start,
			End:   timeRange.End,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to search spans: %w", err)
	}

	// Convert SpanInfo to Span (just pass the map as-is)
	var outputSpans []output.Span
	for _, span := range spans {
		outputSpans = append(outputSpans, output.Span(span))
	}

	// Create spans result
	result := output.SpansResult{
		Spans:   outputSpans,
		WebLink: webLink,
	}

	// Merge with existing data and output
	mergedData := passthroughHandler.MergeSpansResult(existingData, &result)
	return passthroughHandler.OutputMergedResult(mergedData, &result, format)
}
