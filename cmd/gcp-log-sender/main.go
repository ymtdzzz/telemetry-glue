package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

func main() {
	var (
		projectID = flag.String("project-id", "", "GCP Project ID (required)")
		traceID   = flag.String("trace-id", "", "Trace ID to associate with logs (required)")
		message   = flag.String("message", "Test log message from gcp-log-sender", "Log message to send")
		count     = flag.Int("count", 1, "Number of log entries to send")
		severity  = flag.String("severity", "INFO", "Log severity level (DEFAULT, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY)")
	)
	flag.Parse()

	if *projectID == "" {
		log.Fatal("--project-id is required")
	}
	if *traceID == "" {
		log.Fatal("--trace-id is required")
	}

	ctx := context.Background()

	service, err := logging.NewService(ctx, option.WithScopes(logging.LoggingWriteScope))
	if err != nil {
		log.Fatalf("Failed to create logging service: %v", err)
	}

	logName := fmt.Sprintf("projects/%s/logs/gcp-log-sender", *projectID)
	traceResource := fmt.Sprintf("projects/%s/traces/%s", *projectID, *traceID)

	for i := 0; i < *count; i++ {
		entry := &logging.LogEntry{
			LogName: logName,
			Resource: &logging.MonitoredResource{
				Type: "global",
			},
			Severity:    *severity,
			Timestamp:   time.Now().Format(time.RFC3339Nano),
			Trace:       traceResource,
			TextPayload: fmt.Sprintf("%s (entry %d/%d)", *message, i+1, *count),
			Labels: map[string]string{
				"sender": "gcp-log-sender",
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

		fmt.Printf("Successfully sent log entry %d/%d\n", i+1, *count)
		fmt.Printf("  Trace: %s\n", traceResource)
		fmt.Printf("  Message: %s\n", entry.TextPayload)
		fmt.Printf("  Response: %+v\n", resp)

		if i < *count-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	fmt.Printf("\nCompleted sending %d log entries with trace ID: %s\n", *count, *traceID)
	fmt.Printf("You can now test log retrieval using this trace ID in GCP Cloud Logging\n")
}
