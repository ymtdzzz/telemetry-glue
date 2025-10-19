package model

import "time"

// Log represents single log entry
type Log struct {
	Timestamp  time.Time      `json:"timestamp"`
	TraceID    string         `json:"trace_id,omitempty"`
	SpanID     string         `json:"span_id,omitempty"`
	Message    string         `json:"message"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

// Logs represents logs
type Logs []Log
