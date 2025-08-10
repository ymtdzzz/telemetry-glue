package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/gcp"
	"github.com/ymtdzzz/telemetry-glue/cmd/telemetry-glue/newrelic"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "telemetry-glue",
	Short: "A unified interface for observability backends",
	Long: `Telemetry Glue provides backend-specific commands to query data from various
observability backends like New Relic, Google Cloud, Datadog, and others.

Each backend has its own subcommand with vendor-specific arguments and concepts,
while maintaining consistent output formats for downstream processing.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Add backend-specific subcommands
	rootCmd.AddCommand(newrelic.NewRelicCmd())
	rootCmd.AddCommand(gcp.GCPCmd())
}
