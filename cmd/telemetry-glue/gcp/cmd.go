package gcp

import (
	"github.com/spf13/cobra"
)

// GCPCmd creates the gcp subcommand
func GCPCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gcp",
		Short: "GCP Cloud Logging backend commands",
		Long: `Commands for querying log data from GCP Cloud Logging.
These commands use GCP-specific concepts like projects and trace IDs.`,
	}

	// Add subcommands
	cmd.AddCommand(LogsCmd())

	return cmd
}
