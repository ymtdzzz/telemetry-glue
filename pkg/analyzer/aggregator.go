package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ymtdzzz/telemetry-glue/pkg/output"
)

// CombinedData represents the aggregated telemetry data from multiple sources
type CombinedData struct {
	Spans  []output.Span         `json:"spans,omitempty"`
	Logs   []output.LogEntry     `json:"logs,omitempty"`
	Traces []output.TraceSummary `json:"traces,omitempty"`
	Values []string              `json:"values,omitempty"`
}

// DataAggregator handles reading and combining JSON data from stdin
type DataAggregator struct {
	combined *CombinedData
}

// NewDataAggregator creates a new data aggregator
func NewDataAggregator() *DataAggregator {
	return &DataAggregator{
		combined: &CombinedData{},
	}
}

// ReadFromStdin reads JSON objects from stdin and combines them
func (da *DataAggregator) ReadFromStdin(r io.Reader) error {
	// Read all data from stdin
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read from stdin: %w", err)
	}

	// Trim whitespace
	data = []byte(strings.TrimSpace(string(data)))

	// If no data, that's okay
	if len(data) == 0 {
		return nil
	}

	// Try to parse as a single JSON object first
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		return da.addJSONObject(obj)
	}

	// If that fails, try to parse as line-delimited JSON
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	hasContent := false

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		hasContent = true
		var lineObj map[string]interface{}
		if err := json.Unmarshal([]byte(line), &lineObj); err != nil {
			return fmt.Errorf("failed to parse JSON line: %w", err)
		}

		if err := da.addJSONObject(lineObj); err != nil {
			return fmt.Errorf("failed to add JSON object: %w", err)
		}
	}

	// If there was no content, that's okay - just return nil
	if !hasContent {
		return nil
	}

	return scanner.Err()
}

// addJSONObject adds a JSON object to the combined data based on its type
func (da *DataAggregator) addJSONObject(obj map[string]interface{}) error {
	// Check if it's a SpansResult
	if spans, exists := obj["spans"]; exists {
		if err := da.addSpans(spans); err != nil {
			return fmt.Errorf("failed to add spans: %w", err)
		}
	}

	// Check if it's a LogsResult
	if logs, exists := obj["logs"]; exists {
		if err := da.addLogs(logs); err != nil {
			return fmt.Errorf("failed to add logs: %w", err)
		}
	}

	// Check if it's a TopTracesResult
	if traces, exists := obj["traces"]; exists {
		if err := da.addTraces(traces); err != nil {
			return fmt.Errorf("failed to add traces: %w", err)
		}
	}

	// Check if it's a SearchValuesResult
	if values, exists := obj["values"]; exists {
		if err := da.addValues(values); err != nil {
			return fmt.Errorf("failed to add values: %w", err)
		}
	}

	return nil
}

// addSpans adds spans data to the combined data
func (da *DataAggregator) addSpans(spans interface{}) error {
	spansBytes, err := json.Marshal(spans)
	if err != nil {
		return err
	}

	var spansList []output.Span
	if err := json.Unmarshal(spansBytes, &spansList); err != nil {
		return err
	}

	da.combined.Spans = append(da.combined.Spans, spansList...)
	return nil
}

// addLogs adds logs data to the combined data
func (da *DataAggregator) addLogs(logs interface{}) error {
	logsBytes, err := json.Marshal(logs)
	if err != nil {
		return err
	}

	var logsList []output.LogEntry
	if err := json.Unmarshal(logsBytes, &logsList); err != nil {
		return err
	}

	da.combined.Logs = append(da.combined.Logs, logsList...)
	return nil
}

// addTraces adds traces data to the combined data
func (da *DataAggregator) addTraces(traces interface{}) error {
	tracesBytes, err := json.Marshal(traces)
	if err != nil {
		return err
	}

	var tracesList []output.TraceSummary
	if err := json.Unmarshal(tracesBytes, &tracesList); err != nil {
		return err
	}

	da.combined.Traces = append(da.combined.Traces, tracesList...)
	return nil
}

// addValues adds search values data to the combined data
func (da *DataAggregator) addValues(values interface{}) error {
	valuesBytes, err := json.Marshal(values)
	if err != nil {
		return err
	}

	var valuesList []string
	if err := json.Unmarshal(valuesBytes, &valuesList); err != nil {
		return err
	}

	da.combined.Values = append(da.combined.Values, valuesList...)
	return nil
}

// GetCombinedData returns the aggregated data
func (da *DataAggregator) GetCombinedData() *CombinedData {
	return da.combined
}

// Summary returns a summary of the aggregated data
func (cd *CombinedData) Summary() string {
	var parts []string

	if len(cd.Spans) > 0 {
		parts = append(parts, fmt.Sprintf("%d spans", len(cd.Spans)))
	}
	if len(cd.Logs) > 0 {
		parts = append(parts, fmt.Sprintf("%d logs", len(cd.Logs)))
	}
	if len(cd.Traces) > 0 {
		parts = append(parts, fmt.Sprintf("%d traces", len(cd.Traces)))
	}
	if len(cd.Values) > 0 {
		parts = append(parts, fmt.Sprintf("%d values", len(cd.Values)))
	}

	if len(parts) == 0 {
		return "no data"
	}

	return strings.Join(parts, ", ")
}

// GetTimeRange returns the time range covered by the data
func (cd *CombinedData) GetTimeRange() (time.Time, time.Time) {
	var earliest, latest time.Time

	// Check spans for timestamps
	for _, span := range cd.Spans {
		if ts, ok := span["timestamp"].(float64); ok {
			t := time.Unix(int64(ts/1000), 0)
			if earliest.IsZero() || t.Before(earliest) {
				earliest = t
			}
			if latest.IsZero() || t.After(latest) {
				latest = t
			}
		}
	}

	// Check logs for timestamps
	for _, log := range cd.Logs {
		if earliest.IsZero() || log.Timestamp.Before(earliest) {
			earliest = log.Timestamp
		}
		if latest.IsZero() || log.Timestamp.After(latest) {
			latest = log.Timestamp
		}
	}

	// Check traces for timestamps
	for _, trace := range cd.Traces {
		if earliest.IsZero() || trace.StartTime.Before(earliest) {
			earliest = trace.StartTime
		}
		endTime := trace.StartTime.Add(time.Duration(trace.Duration * float64(time.Second)))
		if latest.IsZero() || endTime.After(latest) {
			latest = endTime
		}
	}

	return earliest, latest
}
