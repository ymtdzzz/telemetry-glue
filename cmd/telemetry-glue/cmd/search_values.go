package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/internal/backend"
	"github.com/ymtdzzz/telemetry-glue/pkg/telemetryglue"
)

// searchValuesCmd represents the search-values command
var searchValuesCmd = &cobra.Command{
	Use:   "search-values",
	Short: "Search for unique values of a specified attribute",
	Long: `Search for unique values of a specified attribute across spans.
The query supports wildcard patterns using asterisks (*).

Examples:
  # Search for all paths containing "user"
  telemetry-glue search-values --backend newrelic --attribute http.path --query "*user*"
  
  # Search for all service names
  telemetry-glue search-values --backend newrelic --attribute service.name --query "*"`,
	RunE: runSearchValues,
}

var (
	attribute string
	query     string
	since     string
	until     string
	format    string
)

func runSearchValues(cmd *cobra.Command, args []string) error {
	// Get backend from parent flag
	backendType, _ := cmd.Flags().GetString("backend")

	// Parse time range
	timeRange, err := parseTimeRange(since, until)
	if err != nil {
		return fmt.Errorf("failed to parse time range: %w", err)
	}

	// Create client
	client, err := telemetryglue.NewClient(backendType)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// Create request
	req := backend.SearchValuesRequest{
		Attribute: attribute,
		Query:     query,
		TimeRange: timeRange,
	}

	// Execute search using quick method (single backend)
	resp, err := client.QuickSearchValues(req)
	if err != nil {
		return fmt.Errorf("failed to search values: %w", err)
	}

	// Output results
	return outputSearchValuesResult(resp, format)
}

func parseTimeRange(since, until string) (backend.TimeRange, error) {
	now := time.Now()

	// Parse since
	sinceDuration, err := time.ParseDuration(since)
	if err != nil {
		return backend.TimeRange{}, fmt.Errorf("invalid since duration: %w", err)
	}
	startTime := now.Add(-sinceDuration)

	// Parse until (optional)
	var endTime time.Time
	if until != "" {
		untilDuration, err := time.ParseDuration(until)
		if err != nil {
			return backend.TimeRange{}, fmt.Errorf("invalid until duration: %w", err)
		}
		endTime = now.Add(-untilDuration)
	} else {
		endTime = now
	}

	return backend.TimeRange{
		Start: startTime,
		End:   endTime,
	}, nil
}

func outputSearchValuesResult(resp backend.SearchValuesResponse, format string) error {
	switch format {
	case "json":
		data, err := json.MarshalIndent(resp, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	case "table", "":
		fmt.Printf("Found %d unique values:\n", len(resp.Values))
		for _, value := range resp.Values {
			fmt.Printf("  %s\n", value)
		}
		if resp.WebLink != "" {
			fmt.Printf("\nView in UI: %s\n", resp.WebLink)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(searchValuesCmd)

	// Required flags
	searchValuesCmd.Flags().StringVarP(&attribute, "attribute", "a", "", "Attribute to search (required)")
	searchValuesCmd.Flags().StringVarP(&query, "query", "q", "*", "Search query pattern (supports wildcards)")
	searchValuesCmd.Flags().StringVarP(&since, "since", "s", "1h", "Time range start (e.g., 1h, 30m, 24h)")

	// Optional flags
	searchValuesCmd.Flags().StringVarP(&until, "until", "u", "", "Time range end (default: now)")
	searchValuesCmd.Flags().StringVarP(&format, "format", "f", "table", "Output format (table, json)")

	// Mark required flags
	searchValuesCmd.MarkFlagRequired("attribute")
}
