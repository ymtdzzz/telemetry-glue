package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "telemetry-glue",
	Short: "A unified interface for observability backends",
	Long: `Telemetry Glue provides a unified interface to query data from various
observability backends like New Relic, Datadog, and others.

You can search for attribute values, get top traces, list spans, and more
across different backends using the same command-line interface.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringP("backend", "b", "", "Backend to use (required)")
	rootCmd.MarkPersistentFlagRequired("backend")
}
