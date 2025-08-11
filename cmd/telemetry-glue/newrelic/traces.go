package newrelic

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
)

// TracesFlags holds NewRelic-specific flags for traces command
type TracesFlags struct {
	Common common.CommonFlags
	Entity string
	Field  string
	Value  string
	Limit  int
}

// TracesCmd creates the traces subcommand for NewRelic
func TracesCmd() *cobra.Command {
	flags := &TracesFlags{}

	cmd := &cobra.Command{
		Use:   "traces",
		Short: "Find top traces containing a specific field value in NewRelic",
		Long: `Find top traces containing a specific field value in NewRelic.
This command searches for traces where the specified field exactly matches the given value
and returns the top traces ordered by duration (longest first).

Examples:
  # Find top traces for a specific HTTP path
  telemetry-glue newrelic traces --entity my-app --field http.path --value "/api/users"
  
  # Find top traces for a specific service with custom limit
  telemetry-glue newrelic traces --entity my-app --field service.name --value "user-service" --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTraces(flags)
		},
	}

	// Add NewRelic-specific flags
	cmd.Flags().StringVarP(&flags.Entity, "entity", "e", "", "NewRelic entity name or GUID (required)")
	cmd.Flags().StringVar(&flags.Field, "field", "", "Field to filter by (required)")
	cmd.Flags().StringVarP(&flags.Value, "value", "v", "", "Exact value to match (required)")
	cmd.Flags().IntVarP(&flags.Limit, "limit", "l", 5, "Number of top traces to return (default: 5)")

	// Add common flags
	common.AddCommonFlags(cmd, &flags.Common)

	// Mark required flags
	if err := cmd.MarkFlagRequired("entity"); err != nil {
		panic(fmt.Sprintf("Failed to mark entity flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("field"); err != nil {
		panic(fmt.Sprintf("Failed to mark field flag as required: %v", err))
	}
	if err := cmd.MarkFlagRequired("value"); err != nil {
		panic(fmt.Sprintf("Failed to mark value flag as required: %v", err))
	}

	return cmd
}

func runTraces(flags *TracesFlags) error {
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

	// Execute top traces search
	traces, webLink, err := client.SearchTopTraces(newrelic.TopTracesRequest{
		Entity:    flags.Entity,
		Attribute: flags.Field,
		Value:     flags.Value,
		Limit:     flags.Limit,
		TimeRange: newrelic.TimeRange{
			Start: timeRange.Start,
			End:   timeRange.End,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to search top traces: %w", err)
	}

	// Convert TraceInfo to TraceSummary
	var traceSummaries []output.TraceSummary
	for _, trace := range traces {
		traceSummaries = append(traceSummaries, output.TraceSummary{
			TraceID:   trace.TraceID,
			StartTime: trace.StartTime,
			Duration:  trace.Duration / 1000.0, // Convert ms to seconds
			Attributes: map[string]interface{}{
				"service.name": trace.ServiceName,
				"span_count":   trace.SpanCount,
			},
		})
	}

	// Output results
	result := output.TopTracesResult{
		Traces:  traceSummaries,
		WebLink: webLink,
	}

	return result.Print(format)
}
