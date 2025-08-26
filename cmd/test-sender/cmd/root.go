package cmd

import (
	"os"

	"github.com/spf13/cobra"
	logcmd "github.com/ymtdzzz/telemetry-glue/cmd/test-sender/cmd/log"
	tracecmd "github.com/ymtdzzz/telemetry-glue/cmd/test-sender/cmd/trace"
)

var rootCmd = &cobra.Command{
	Use:   "test-sender",
	Short: "Send test telemetry data to various backends",
	Long: `Test Sender is a CLI tool for sending test telemetry data (logs, traces, metrics)
to various observability backends like GCP Cloud Logging, New Relic, and others.

Use this tool to generate test data for development and testing purposes.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(logcmd.LogCmd())
	rootCmd.AddCommand(tracecmd.TraceCmd())
}
