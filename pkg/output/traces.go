package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Trace represents a trace summary for output
type Trace struct {
	TraceID     string    `json:"trace_id"`
	Duration    float64   `json:"duration_millis"`
	ServiceName string    `json:"service_name"`
	SpanCount   int       `json:"span_count"`
	StartTime   time.Time `json:"start_time"`
}

// TracesResult represents top traces output
type TracesResult struct {
	Traces []Trace `json:"traces"`
}

// Print outputs top traces result in the specified format
func (r TracesResult) Print(format Format) error {
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

func (r TracesResult) printJSON() error {
	return printJSON(r)
}

func (r TracesResult) printTable() error {
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

func (r TracesResult) printCSV() error {
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
