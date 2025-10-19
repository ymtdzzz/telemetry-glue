package model

import "time"

// Telemetry represents telemetry data including spans and logs
type Telemetry struct {
	Spans Spans
	Logs  Logs
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
