package backend

import (
	"context"

	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
)

// SearchLogsRequest represents a request to search spans
type SearchSpansRequest struct {
	TraceID   string
	TimeRange *TimeRange
}

// SearchLogsRequest represents a request to search logs
type SearchLogsRequest struct {
	TraceID   string
	TimeRange *TimeRange
}

// GlueBackend defines the interface for backends that can search both spans and logs
type GlueBackend interface {
	SearchSpans(ctx context.Context, req *SearchSpansRequest) (model.Spans, error)
	SearchLogs(ctx context.Context, req *SearchLogsRequest) (model.Logs, error)
}
