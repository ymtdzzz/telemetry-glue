package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Format represents supported output formats
type Format string

const (
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
	FormatTable Format = "table"
)

// SearchValuesResult represents search values output
type SearchValuesResult struct {
	Values  []string `json:"values"`
	WebLink string   `json:"web_link,omitempty"`
}

// TraceSummary represents a trace summary for output
type TraceSummary struct {
	TraceID    string                 `json:"trace_id"`
	StartTime  time.Time              `json:"start_time"`
	Duration   float64                `json:"duration_seconds"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// TopTracesResult represents top traces output
type TopTracesResult struct {
	Traces  []TraceSummary `json:"traces"`
	WebLink string         `json:"web_link,omitempty"`
}

// Span represents a span for output
type Span map[string]interface{}

// SpansResult represents spans output
type SpansResult struct {
	Spans   []Span `json:"spans"`
	WebLink string `json:"web_link,omitempty"`
}

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
	Logs    []LogEntry `json:"logs"`
	WebLink string     `json:"web_link,omitempty"`
}

// PrintSearchValues outputs search values result in the specified format
func PrintSearchValues(result SearchValuesResult, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(result)
	case FormatCSV:
		return printSearchValuesCSV(result)
	case FormatTable, "":
		return printSearchValuesTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// PrintTopTraces outputs top traces result in the specified format
func PrintTopTraces(result TopTracesResult, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(result)
	case FormatCSV:
		return printTopTracesCSV(result)
	case FormatTable, "":
		return printTopTracesTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// PrintSpans outputs spans result in the specified format
func PrintSpans(result SpansResult, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(result)
	case FormatCSV:
		return printSpansCSV(result)
	case FormatTable, "":
		return printSpansTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// PrintLogs outputs logs result in the specified format
func PrintLogs(result LogsResult, format Format) error {
	switch format {
	case FormatJSON:
		return printJSON(result)
	case FormatCSV:
		return printLogsCSV(result)
	case FormatTable, "":
		return printLogsTable(result)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func printSearchValuesTable(result SearchValuesResult) error {
	fmt.Printf("Found %d unique values:\n", len(result.Values))
	for _, value := range result.Values {
		fmt.Printf("  %s\n", value)
	}
	if result.WebLink != "" {
		fmt.Printf("\nView in UI: %s\n", result.WebLink)
	}
	return nil
}

func printSearchValuesCSV(result SearchValuesResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"value"}); err != nil {
		return err
	}

	// Write values
	for _, value := range result.Values {
		if err := writer.Write([]string{value}); err != nil {
			return err
		}
	}

	// Print web link separately for CSV
	if result.WebLink != "" {
		fmt.Fprintf(os.Stderr, "# View in UI: %s\n", result.WebLink)
	}

	return nil
}

func printTopTracesTable(result TopTracesResult) error {
	fmt.Printf("Top %d traces:\n", len(result.Traces))
	for i, trace := range result.Traces {
		fmt.Printf("%d. %s (%s) - %.3fs\n",
			i+1,
			trace.TraceID,
			trace.StartTime.Format("2006-01-02 15:04:05"),
			trace.Duration)
	}
	if result.WebLink != "" {
		fmt.Printf("\nView in UI: %s\n", result.WebLink)
	}
	return nil
}

func printTopTracesCSV(result TopTracesResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"trace_id", "start_time", "duration_seconds"}); err != nil {
		return err
	}

	// Write traces
	for _, trace := range result.Traces {
		if err := writer.Write([]string{
			trace.TraceID,
			trace.StartTime.Format(time.RFC3339),
			strconv.FormatFloat(trace.Duration, 'f', 3, 64),
		}); err != nil {
			return err
		}
	}

	// Print web link separately for CSV
	if result.WebLink != "" {
		fmt.Fprintf(os.Stderr, "# View in UI: %s\n", result.WebLink)
	}

	return nil
}

func printSpansTable(result SpansResult) error {
	fmt.Printf("Found %d spans:\n\n", len(result.Spans))

	if len(result.Spans) == 0 {
		fmt.Println("No spans found.")
		return nil
	}

	for i, span := range result.Spans {
		fmt.Printf("=== Span %d ===\n", i+1)

		// Sort keys for consistent output
		var keys []string
		for key := range span {
			keys = append(keys, key)
		}

		// Simple alphabetical sort
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[i] > keys[j] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}

		// Print all key-value pairs
		for _, key := range keys {
			value := span[key]

			// Format different types appropriately
			var valueStr string
			switch v := value.(type) {
			case string:
				valueStr = v
			case float64:
				// Special formatting for timestamp
				if key == "timestamp" {
					valueStr = fmt.Sprintf("%.0f (%s)", v, time.Unix(int64(v/1000), 0).Format("2006-01-02 15:04:05"))
				} else if key == "duration.ms" {
					valueStr = fmt.Sprintf("%.3f ms", v)
				} else {
					valueStr = fmt.Sprintf("%.6g", v)
				}
			case bool:
				valueStr = fmt.Sprintf("%t", v)
			case nil:
				valueStr = "<nil>"
			default:
				valueStr = fmt.Sprintf("%v", v)
			}

			// Truncate very long values but show they're truncated
			if len(valueStr) > 100 {
				valueStr = valueStr[:97] + "..."
			}

			fmt.Printf("  %-30s: %s\n", key, valueStr)
		}

		if i < len(result.Spans)-1 {
			fmt.Println()
		}
	}

	if result.WebLink != "" {
		fmt.Printf("\nView in UI: %s\n", result.WebLink)
	}
	return nil
}
func printSpansCSV(result SpansResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"span_id", "trace_id", "name", "parent_id", "timestamp", "duration_ms", "service_name", "operation", "resource"}); err != nil {
		return err
	}

	// Write spans
	for _, span := range result.Spans {
		// Extract values with type assertions and provide defaults using correct field names
		spanID := ""
		if id, ok := span["id"].(string); ok {
			spanID = id
		}

		traceID := ""
		if id, ok := span["trace.id"].(string); ok {
			traceID = id
		}

		name := ""
		if n, ok := span["name"].(string); ok {
			name = n
		}

		parentID := ""
		if pid, ok := span["parent.id"].(string); ok {
			parentID = pid
		}

		timestamp := ""
		if ts, ok := span["timestamp"].(float64); ok {
			timestamp = time.Unix(int64(ts/1000), 0).Format(time.RFC3339Nano)
		}

		duration := ""
		if d, ok := span["duration.ms"].(float64); ok {
			duration = strconv.FormatFloat(d, 'f', 3, 64)
		}

		serviceName := ""
		if sn, ok := span["service.name"].(string); ok {
			serviceName = sn
		}

		operation := ""
		if op, ok := span["operation.name"].(string); ok {
			operation = op
		}

		resource := ""
		if res, ok := span["resource.name"].(string); ok {
			resource = res
		}

		if err := writer.Write([]string{
			spanID,
			traceID,
			name,
			parentID,
			timestamp,
			duration,
			serviceName,
			operation,
			resource,
		}); err != nil {
			return err
		}
	}

	// Print web link separately for CSV
	if result.WebLink != "" {
		fmt.Fprintf(os.Stderr, "# View in UI: %s\n", result.WebLink)
	}

	return nil
}

func printLogsTable(result LogsResult) error {
	fmt.Printf("Found %d log entries:\n", len(result.Logs))
	for _, log := range result.Logs {
		fmt.Printf("  %s: %s\n",
			log.Timestamp.Format("2006-01-02 15:04:05.000"),
			truncateString(log.Message, 80))
	}
	if result.WebLink != "" {
		fmt.Printf("\nView in UI: %s\n", result.WebLink)
	}
	return nil
}

func printLogsCSV(result LogsResult) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"timestamp", "trace_id", "span_id", "message"}); err != nil {
		return err
	}

	// Write logs
	for _, log := range result.Logs {
		if err := writer.Write([]string{
			log.Timestamp.Format(time.RFC3339Nano),
			log.TraceID,
			log.SpanID,
			log.Message,
		}); err != nil {
			return err
		}
	}

	// Print web link separately for CSV
	if result.WebLink != "" {
		fmt.Fprintf(os.Stderr, "# View in UI: %s\n", result.WebLink)
	}

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ParseFormat parses a format string into a Format enum
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "table", "":
		return FormatTable, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: json, csv, table)", s)
	}
}
