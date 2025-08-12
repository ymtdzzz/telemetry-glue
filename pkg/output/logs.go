package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// LogEntry represents a log entry for output
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	TraceID    string                 `json:"trace_id,omitempty"`
	SpanID     string                 `json:"span_id,omitempty"`
	Message    string                 `json:"message"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// LogsResult represents logs output
type LogsResult struct {
	Logs []LogEntry `json:"logs"`
}

// Print outputs logs result in the specified format
func (r LogsResult) Print(format Format) error {
	switch format {
	case FormatJSON:
		return r.printJSON()
	case FormatCSV:
		return r.printCSV()
	case FormatTable, "":
		return r.printTable()
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func (r LogsResult) printJSON() error {
	return printJSON(r)
}

func (r LogsResult) printTable() error {
	fmt.Printf("Found %d log entries:\n", len(r.Logs))
	for _, log := range r.Logs {
		fmt.Printf("  %s: %s\n",
			log.Timestamp.Format("2006-01-02 15:04:05.000"),
			truncateString(log.Message, 80))
	}

	return nil
}

func (r LogsResult) printCSV() error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"timestamp", "trace_id", "span_id", "message"}); err != nil {
		return err
	}

	// Write logs
	for _, log := range r.Logs {
		if err := writer.Write([]string{
			log.Timestamp.Format(time.RFC3339Nano),
			log.TraceID,
			log.SpanID,
			log.Message,
		}); err != nil {
			return err
		}
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
