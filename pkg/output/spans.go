package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

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
