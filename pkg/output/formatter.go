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

// Formatter interface for types that can format their output
type Formatter interface {
	Print(format Format) error
}

// SearchValuesResult represents search values output
type SearchValuesResult struct {
	Values []string `json:"values"`
}

// Print outputs search values result in the specified format
func (r SearchValuesResult) Print(format Format) error {
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

func (r SearchValuesResult) printJSON() error {
	return printJSON(r)
}

func (r SearchValuesResult) printTable() error {
	fmt.Printf("Found %d unique values:\n", len(r.Values))
	for _, value := range r.Values {
		fmt.Printf("  %s\n", value)
	}

	return nil
}

func (r SearchValuesResult) printCSV() error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"value"}); err != nil {
		return err
	}

	// Write values
	for _, value := range r.Values {
		if err := writer.Write([]string{value}); err != nil {
			return err
		}
	}



	return nil
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
	Traces []TraceSummary `json:"traces"`
}

// Print outputs top traces result in the specified format
func (r TopTracesResult) Print(format Format) error {
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

func (r TopTracesResult) printJSON() error {
	return printJSON(r)
}

func (r TopTracesResult) printTable() error {
	fmt.Printf("Top %d traces:\n", len(r.Traces))
	for i, trace := range r.Traces {
		fmt.Printf("%d. %s (%s) - %.3fs\n",
			i+1,
			trace.TraceID,
			trace.StartTime.Format("2006-01-02 15:04:05"),
			trace.Duration)
	}

	return nil
}

func (r TopTracesResult) printCSV() error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"trace_id", "start_time", "duration_seconds"}); err != nil {
		return err
	}

	// Write traces
	for _, trace := range r.Traces {
		if err := writer.Write([]string{
			trace.TraceID,
			trace.StartTime.Format(time.RFC3339),
			strconv.FormatFloat(trace.Duration, 'f', 3, 64),
		}); err != nil {
			return err
		}
	}



	return nil
}

// Span represents a span for output
type Span map[string]interface{}

// SpansResult represents spans output
type SpansResult struct {
	Spans []Span `json:"spans"`
}

// Print outputs spans result in the specified format
func (r SpansResult) Print(format Format) error {
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

func (r SpansResult) printJSON() error {
	return printJSON(r)
}

func (r SpansResult) printTable() error {
	fmt.Printf("Found %d spans:\n\n", len(r.Spans))

	if len(r.Spans) == 0 {
		fmt.Println("No spans found.")
		return nil
	}

	for i, span := range r.Spans {
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
				switch key {
				case "timestamp":
					valueStr = fmt.Sprintf("%.0f (%s)", v, time.Unix(int64(v/1000), 0).Format("2006-01-02 15:04:05"))
				case "duration.ms":
					valueStr = fmt.Sprintf("%.3f ms", v)
				default:
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

		if i < len(r.Spans)-1 {
			fmt.Println()
		}
	}


	return nil
}

func (r SpansResult) printCSV() error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	if err := writer.Write([]string{"span_id", "trace_id", "name", "parent_id", "timestamp", "duration_ms", "service_name", "operation", "resource"}); err != nil {
		return err
	}

	// Write spans
	for _, span := range r.Spans {
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



	return nil
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

func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
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
