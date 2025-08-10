package newrelic

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/common"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
)

// SearchValuesFlags holds NewRelic-specific flags for search-values command
type SearchValuesFlags struct {
	Common    common.CommonFlags
	Entity    string
	Attribute string
	Query     string
}

// SearchValuesCmd creates the search-values subcommand for NewRelic
func SearchValuesCmd() *cobra.Command {
	flags := &SearchValuesFlags{}

	cmd := &cobra.Command{
		Use:   "search-values",
		Short: "Search for unique values of a specified attribute in NewRelic",
		Long: `Search for unique values of a specified attribute across spans in NewRelic.
The query supports wildcard patterns using asterisks (*).

Examples:
  # Search for all paths containing "user" in entity "my-app"
  telemetry-glue newrelic search-values --entity my-app --attribute http.path --query "*user*"
  
  # Search for all service names
  telemetry-glue newrelic search-values --entity my-app --attribute service.name --query "*"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearchValues(flags)
		},
	}

	// Add NewRelic-specific flags
	cmd.Flags().StringVarP(&flags.Entity, "entity", "e", "", "NewRelic entity name or GUID (required)")
	cmd.Flags().StringVarP(&flags.Attribute, "attribute", "a", "", "Attribute to search (required)")
	cmd.Flags().StringVarP(&flags.Query, "query", "q", "*", "Search query pattern (supports wildcards)")

	// Add common flags
	common.AddCommonFlags(cmd, &flags.Common)

	// Mark required flags
	cmd.MarkFlagRequired("entity")
	cmd.MarkFlagRequired("attribute")

	return cmd
}

func runSearchValues(flags *SearchValuesFlags) error {
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

	// Execute search
	values, webLink, err := client.SearchValues(newrelic.SearchValuesRequest{
		Entity:    flags.Entity,
		Attribute: flags.Attribute,
		Query:     flags.Query,
		TimeRange: newrelic.TimeRange{
			Start: timeRange.Start,
			End:   timeRange.End,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to search values: %w", err)
	}

	// Output results
	result := output.SearchValuesResult{
		Values:  values,
		WebLink: webLink,
	}

	return result.Print(format)
}
