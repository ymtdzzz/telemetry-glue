package backend

import "time"

// TimeRange represents a time range for queries.
type TimeRange struct {
	Start time.Time // Start of the range
	End   time.Time // End of the range
}

// SearchValuesRequest represents a request to search for attribute values.
type SearchValuesRequest struct {
	Attribute string // e.g. "http.path"
	Query     string // e.g. "*user*"
	TimeRange TimeRange
}

// SearchValuesResponse represents the result of a search values query.
type SearchValuesResponse struct {
	Values  []string
	WebLink string // Link to the relevant search result in the backend UI
}

// TopTracesRequest represents a request to get top traces.
type TopTracesRequest struct {
	Attribute string // e.g. "http.path"
	Value     string // e.g. "/admin/users/new"
	TimeRange TimeRange
	Limit     int
}

// TopTracesResponse represents the result of a top traces query.
type TopTracesResponse struct {
	Traces  []TraceSummary
	WebLink string // Link to the search result in the backend UI
}

// ListSpansRequest represents a request to list spans for a trace.
type ListSpansRequest struct {
	TraceID   string
	TimeRange TimeRange
}

// ListSpansResponse represents the result of a list spans query.
type ListSpansResponse struct {
	Spans   []Span
	WebLink string // Link to the trace in the backend UI
}

// ListLogsRequest represents a request to list logs for a trace.
type ListLogsRequest struct {
	TraceID   string
	TimeRange TimeRange
}

// ListLogsResponse represents the result of a list logs query.
type ListLogsResponse struct {
	Logs    []LogEntry
	WebLink string // Link to the logs in the backend UI
}

// TraceSummary, Span, LogEntry are domain models to be defined as needed.
type TraceSummary struct {
	TraceID    string
	StartTime  time.Time
	Duration   float64 // seconds
	Attributes map[string]any
}

type Span struct {
	SpanID     string
	TraceID    string
	Name       string
	StartTime  time.Time
	EndTime    time.Time
	Attributes map[string]any
}

type LogEntry struct {
	Timestamp  time.Time
	TraceID    string
	SpanID     string
	Message    string
	Attributes map[string]any
}

// Backend is the interface that all observability backends must implement.
type Backend interface {
	Name() string

	SearchValues(req SearchValuesRequest) (SearchValuesResponse, error)
	TopTraces(req TopTracesRequest) (TopTracesResponse, error)
	ListSpans(req ListSpansRequest) (ListSpansResponse, error)
	ListLogs(req ListLogsRequest) (ListLogsResponse, error)
}
