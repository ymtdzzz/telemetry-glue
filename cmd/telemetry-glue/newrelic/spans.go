package newrelic

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
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
  telemetry-glue newrelic spans --trace-id d8a60536187fa0927e45911f8c0dd64b --since 2h`,
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
	// Parse time range
	timeRange, err := common.ParseTimeRange(flags.Common.Since, flags.Common.Until)
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

	// Output results
	result := output.SpansResult{
		Spans:   outputSpans,
		WebLink: webLink,
	}

	return output.PrintSpans(result, format)
}
