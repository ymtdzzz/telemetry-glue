package newrelic

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
)

// TopTracesFlags holds NewRelic-specific flags for top-traces command
type TopTracesFlags struct {
	Common    common.CommonFlags
	Entity    string
	Attribute string
	Value     string
	Limit     int
}

// TopTracesCmd creates the top-traces subcommand for NewRelic
func TopTracesCmd() *cobra.Command {
	flags := &TopTracesFlags{}

	cmd := &cobra.Command{
		Use:   "top-traces",
		Short: "Find top traces containing a specific attribute value in NewRelic",
		Long: `Find top traces containing a specific attribute value in NewRelic.
This command searches for traces where the specified attribute exactly matches the given value
and returns the top traces ordered by duration (longest first).

Examples:
  # Find top traces for a specific HTTP path
  telemetry-glue newrelic top-traces --entity my-app --attribute http.path --value "/api/users"
  
  # Find top traces for a specific service with custom limit
  telemetry-glue newrelic top-traces --entity my-app --attribute service.name --value "user-service" --limit 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTopTraces(flags)
		},
	}

	// Add NewRelic-specific flags
	cmd.Flags().StringVarP(&flags.Entity, "entity", "e", "", "NewRelic entity name or GUID (required)")
	cmd.Flags().StringVarP(&flags.Attribute, "attribute", "a", "", "Attribute to filter by (required)")
	cmd.Flags().StringVarP(&flags.Value, "value", "v", "", "Exact value to match (required)")
	cmd.Flags().IntVarP(&flags.Limit, "limit", "l", 5, "Number of top traces to return (default: 5)")

	// Add common flags
	common.AddCommonFlags(cmd, &flags.Common)

	// Mark required flags
	cmd.MarkFlagRequired("entity")
	cmd.MarkFlagRequired("attribute")
	cmd.MarkFlagRequired("value")

	return cmd
}

func runTopTraces(flags *TopTracesFlags) error {
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

	// Execute top traces search
	traces, webLink, err := client.SearchTopTraces(newrelic.TopTracesRequest{
		Entity:    flags.Entity,
		Attribute: flags.Attribute,
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
