package trace

import (
	"github.com/spf13/cobra"
)

var traceCmd = &cobra.Command{
	Use:   "trace",
	Short: "Send test traces",
	Long:  `Send test traces to various tracing backends`,
}

func TraceCmd() *cobra.Command {
	return traceCmd
}

func init() {
	traceCmd.AddCommand(newrelicTraceCmd)
}
