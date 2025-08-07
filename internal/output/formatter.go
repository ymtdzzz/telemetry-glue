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
type Span struct {
	SpanID     string                 `json:"span_id"`
	TraceID    string                 `json:"trace_id"`
	Name       string                 `json:"name"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

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
	fmt.Printf("Found %d spans:\n", len(result.Spans))
	for _, span := range result.Spans {
		duration := span.EndTime.Sub(span.StartTime)
		fmt.Printf("  %s: %s (%s - %s) %.3fs\n",
			span.SpanID,
			span.Name,
			span.StartTime.Format("15:04:05.000"),
			span.EndTime.Format("15:04:05.000"),
			duration.Seconds())
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
	if err := writer.Write([]string{"span_id", "trace_id", "name", "start_time", "end_time", "duration_seconds"}); err != nil {
		return err
	}

	// Write spans
	for _, span := range result.Spans {
		duration := span.EndTime.Sub(span.StartTime)
		if err := writer.Write([]string{
			span.SpanID,
			span.TraceID,
			span.Name,
			span.StartTime.Format(time.RFC3339Nano),
			span.EndTime.Format(time.RFC3339Nano),
			strconv.FormatFloat(duration.Seconds(), 'f', 6, 64),
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
