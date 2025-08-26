package log

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

var gcpLogCmd = &cobra.Command{
	Use:   "gcp",
	Short: "Send test logs to GCP Cloud Logging",
	Long:  `Send test log entries to Google Cloud Logging with specified trace IDs`,
	RunE:  runGCPLog,
}

var (
	gcpProjectID string
	gcpTraceID   string
	gcpMessage   string
	gcpCount     int
	gcpSeverity  string
)

func init() {
	gcpLogCmd.Flags().StringVar(&gcpProjectID, "project-id", "", "GCP Project ID (required)")
	gcpLogCmd.Flags().StringVar(&gcpTraceID, "trace-id", "", "Trace ID to associate with logs (required)")
	gcpLogCmd.Flags().StringVar(&gcpMessage, "message", "Test log message from test-sender", "Log message to send")
	gcpLogCmd.Flags().IntVar(&gcpCount, "count", 1, "Number of log entries to send")
	gcpLogCmd.Flags().StringVar(&gcpSeverity, "severity", "INFO", "Log severity level (DEFAULT, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY)")

	gcpLogCmd.MarkFlagRequired("project-id")
	gcpLogCmd.MarkFlagRequired("trace-id")
}

func runGCPLog(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	service, err := logging.NewService(ctx, option.WithScopes(logging.LoggingWriteScope))
	if err != nil {
		return fmt.Errorf("failed to create logging service: %w", err)
	}

	logName := fmt.Sprintf("projects/%s/logs/test-sender", gcpProjectID)
	traceResource := fmt.Sprintf("projects/%s/traces/%s", gcpProjectID, gcpTraceID)

	for i := 0; i < gcpCount; i++ {
		entry := &logging.LogEntry{
			LogName: logName,
			Resource: &logging.MonitoredResource{
				Type: "global",
			},
			Severity:    gcpSeverity,
			Timestamp:   time.Now().Format(time.RFC3339Nano),
			Trace:       traceResource,
			TextPayload: fmt.Sprintf("%s (entry %d/%d)", gcpMessage, i+1, gcpCount),
			Labels: map[string]string{
				"sender": "test-sender",
			},
		}

		req := &logging.WriteLogEntriesRequest{
			LogName: logName,
			Resource: &logging.MonitoredResource{
				Type: "global",
			},
			Entries: []*logging.LogEntry{entry},
		}

		resp, err := service.Entries.Write(req).Do()
		if err != nil {
			log.Printf("Failed to write log entry %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Successfully sent log entry %d/%d\n", i+1, gcpCount)
		fmt.Printf("  Trace: %s\n", traceResource)
		fmt.Printf("  Message: %s\n", entry.TextPayload)
		fmt.Printf("  Response: %+v\n", resp)

		if i < gcpCount-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Printf("\nCompleted sending %d log entries with trace ID: %s\n", gcpCount, gcpTraceID)
	fmt.Printf("You can now test log retrieval using this trace ID in GCP Cloud Logging\n")

	return nil
}
