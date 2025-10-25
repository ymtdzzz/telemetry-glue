package model

import (
	"fmt"
	"time"
)

// Telemetry represents telemetry data including spans and logs
type Telemetry struct {
	Spans Spans `json:"spans"`
	Logs  Logs  `json:"logs"`
}

func (t *Telemetry) TimeRange() (time.Time, time.Time) {
	var earliest, latest time.Time

	// Check spans for timestamps
	for _, span := range t.Spans {
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

	return earliest, latest
}

// AsCSV converts telemetry data to CSV format
func (t *Telemetry) AsCSV() (spans string, logs string, err error) {
	spans, err = t.Spans.AsCSV()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert spans to CSV: %w", err)
	}

	logs, err = t.Logs.AsCSV()
	if err != nil {
		return "", "", fmt.Errorf("failed to convert logs to CSV: %w", err)
	}

	return
}

// RoughTokenEstimate provides a rough estimate of the token count for the telemetry data
func (t *Telemetry) RoughTokenEstimate() (int, error) {
	spans, logs, err := t.AsCSV()
	if err != nil {
		return 0, fmt.Errorf("failed to convert telemetry to CSV for token estimation: %w", err)
	}

	return len([]rune(string(spans+logs))) / 3, nil
}
