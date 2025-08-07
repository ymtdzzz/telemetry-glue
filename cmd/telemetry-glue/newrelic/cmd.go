package newrelic

import (
	"github.com/spf13/cobra"
)

// NewRelicCmd creates the newrelic subcommand
func NewRelicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "newrelic",
		Short: "NewRelic backend commands",
		Long: `Commands for querying telemetry data from NewRelic.
These commands use NewRelic-specific concepts like entities and accounts.`,
	}

	// Add subcommands
	cmd.AddCommand(SearchValuesCmd())

	return cmd
}
