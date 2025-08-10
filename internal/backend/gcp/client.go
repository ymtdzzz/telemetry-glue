package gcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/api/logging/v2"
	"google.golang.org/api/option"
)

// Common errors for GCP backend
var (
	ErrMissingProjectID = errors.New("project ID is required")
	ErrMissingTraceID   = errors.New("trace ID is required")
)

// LogsRequest represents a request to search for logs in GCP Cloud Logging
type LogsRequest struct {
	ProjectID string    // GCP Project ID
	TraceID   string    // Trace ID to search for
	Limit     int       // Maximum number of log entries to return
	StartTime time.Time // Start time for the search
	EndTime   time.Time // End time for the search
}

// LogEntry represents a log entry from GCP Cloud Logging
type LogEntry struct {
	LogName        string                 `json:"log_name"`
	Resource       map[string]interface{} `json:"resource"`
	Timestamp      time.Time              `json:"timestamp"`
	Severity       string                 `json:"severity"`
	InsertID       string                 `json:"insert_id,omitempty"`
	HTTPRequest    map[string]interface{} `json:"http_request,omitempty"`
	Labels         map[string]string      `json:"labels,omitempty"`
	Operation      map[string]interface{} `json:"operation,omitempty"`
	Trace          string                 `json:"trace,omitempty"`
	SpanID         string                 `json:"span_id,omitempty"`
	TraceSampled   bool                   `json:"trace_sampled,omitempty"`
	SourceLocation map[string]interface{} `json:"source_location,omitempty"`

	// Payload fields (only one will be populated)
	TextPayload  string                 `json:"text_payload,omitempty"`
	JSONPayload  map[string]interface{} `json:"json_payload,omitempty"`
	ProtoPayload map[string]interface{} `json:"proto_payload,omitempty"`
}

// Client represents a GCP Cloud Logging client
type Client struct {
	service *logging.Service
}

// NewClient creates a new GCP Cloud Logging client
func NewClient() (*Client, error) {
	ctx := context.Background()

	// Create logging service with default credentials
	service, err := logging.NewService(ctx, option.WithScopes(logging.LoggingReadScope))
	if err != nil {
		return nil, fmt.Errorf("failed to create logging service: %w", err)
	}

	return &Client{
		service: service,
	}, nil
}

// SearchLogs searches for log entries by trace ID
func (c *Client) SearchLogs(req LogsRequest) ([]LogEntry, string, error) {
	if req.ProjectID == "" {
		return nil, "", ErrMissingProjectID
	}
	if req.TraceID == "" {
		return nil, "", ErrMissingTraceID
	}

	// Build the filter for trace ID
	traceResource := fmt.Sprintf("projects/%s/traces/%s", req.ProjectID, req.TraceID)
	filter := fmt.Sprintf(`trace="%s"`, traceResource)

	// Add time range if specified
	if !req.StartTime.IsZero() {
		filter += fmt.Sprintf(` AND timestamp>="%s"`, req.StartTime.Format(time.RFC3339))
	}
	if !req.EndTime.IsZero() {
		filter += fmt.Sprintf(` AND timestamp<="%s"`, req.EndTime.Format(time.RFC3339))
	}

	// Set default limit if not specified
	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	// Create the request
	listReq := &logging.ListLogEntriesRequest{
		ResourceNames: []string{fmt.Sprintf("projects/%s", req.ProjectID)},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      int64(limit),
	}

	// Execute the request
	resp, err := c.service.Entries.List(listReq).Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list log entries: %w", err)
	}

	// Convert response to our LogEntry format
	var entries []LogEntry
	for _, entry := range resp.Entries {
		logEntry := LogEntry{
			LogName:      entry.LogName,
			Severity:     entry.Severity,
			InsertID:     entry.InsertId,
			Trace:        entry.Trace,
			SpanID:       entry.SpanId,
			TraceSampled: entry.TraceSampled,
		}

		// Parse timestamp
		if entry.Timestamp != "" {
			if ts, err := time.Parse(time.RFC3339Nano, entry.Timestamp); err == nil {
				logEntry.Timestamp = ts
			}
		}

		// Handle resource
		if entry.Resource != nil {
			logEntry.Resource = map[string]interface{}{
				"type":   entry.Resource.Type,
				"labels": entry.Resource.Labels,
			}
		}

		// Handle labels
		if entry.Labels != nil {
			logEntry.Labels = entry.Labels
		}

		// Handle HTTP request
		if entry.HttpRequest != nil {
			logEntry.HTTPRequest = map[string]interface{}{
				"request_method":                     entry.HttpRequest.RequestMethod,
				"request_url":                        entry.HttpRequest.RequestUrl,
				"request_size":                       entry.HttpRequest.RequestSize,
				"status":                             entry.HttpRequest.Status,
				"response_size":                      entry.HttpRequest.ResponseSize,
				"user_agent":                         entry.HttpRequest.UserAgent,
				"remote_ip":                          entry.HttpRequest.RemoteIp,
				"server_ip":                          entry.HttpRequest.ServerIp,
				"referer":                            entry.HttpRequest.Referer,
				"latency":                            entry.HttpRequest.Latency,
				"cache_lookup":                       entry.HttpRequest.CacheLookup,
				"cache_hit":                          entry.HttpRequest.CacheHit,
				"cache_validated_with_origin_server": entry.HttpRequest.CacheValidatedWithOriginServer,
				"cache_fill_bytes":                   entry.HttpRequest.CacheFillBytes,
				"protocol":                           entry.HttpRequest.Protocol,
			}
		}

		// Handle operation
		if entry.Operation != nil {
			logEntry.Operation = map[string]interface{}{
				"id":       entry.Operation.Id,
				"producer": entry.Operation.Producer,
				"first":    entry.Operation.First,
				"last":     entry.Operation.Last,
			}
		}

		// Handle source location
		if entry.SourceLocation != nil {
			logEntry.SourceLocation = map[string]interface{}{
				"file":     entry.SourceLocation.File,
				"line":     entry.SourceLocation.Line,
				"function": entry.SourceLocation.Function,
			}
		}

		// Handle payload (only one should be set)
		if entry.TextPayload != "" {
			logEntry.TextPayload = entry.TextPayload
		} else if entry.JsonPayload != nil {
			// Parse JSON payload
			var jsonData map[string]interface{}
			if err := json.Unmarshal(entry.JsonPayload, &jsonData); err == nil {
				logEntry.JSONPayload = jsonData
			}
		} else if entry.ProtoPayload != nil {
			// Parse proto payload
			var protoData map[string]interface{}
			if err := json.Unmarshal(entry.ProtoPayload, &protoData); err == nil {
				logEntry.ProtoPayload = protoData
			}
		}

		entries = append(entries, logEntry)
	}

	// Generate web link for GCP Cloud Logging console
	webLink := c.generateWebLink(req.ProjectID, req.TraceID)

	return entries, webLink, nil
}

// generateWebLink generates a GCP Cloud Logging console link
func (c *Client) generateWebLink(projectID, traceID string) string {
	// URL encode the filter
	encodedFilter := fmt.Sprintf("trace=%%22projects/%s/traces/%s%%22", projectID, traceID)

	return fmt.Sprintf("https://console.cloud.google.com/logs/query;query=%s?project=%s",
		encodedFilter, projectID)
}
