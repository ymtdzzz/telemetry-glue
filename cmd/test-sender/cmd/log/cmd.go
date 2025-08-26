package log

import (
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Send test log entries",
	Long:  `Send test log entries to various logging backends`,
}

func LogCmd() *cobra.Command {
	return logCmd
}

func init() {
	logCmd.AddCommand(gcpLogCmd)
}
