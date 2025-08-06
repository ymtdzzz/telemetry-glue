// Package telemetryglue provides a unified client library for querying observability backends
package telemetryglue

import (
	"fmt"

	"github.com/ymtdzzz/telemetry-glue/internal/backend"
	"github.com/ymtdzzz/telemetry-glue/internal/backend/newrelic"
	"github.com/ymtdzzz/telemetry-glue/internal/core"
)

// Client represents a telemetry glue client that provides a unified interface
// to various observability backends
type Client struct {
	coreClient *core.Client
}

// NewClient creates a new telemetry glue client with the specified backend
func NewClient(backendType string) (*Client, error) {
	if err := backend.ValidateBackendType(backendType); err != nil {
		return nil, err
	}

	// Create core client
	coreClient, err := core.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create core client: %w", err)
	}

	// Initialize backend and add to core client
	switch backendType {
	case "newrelic":
		nrBackend, err := newrelic.NewNewRelicBackend()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize NewRelic backend: %w", err)
		}
		// Add backend directly (note: interface compatibility will be resolved later)
		coreClient.AddBackend(backendType, nrBackend)
	default:
		return nil, fmt.Errorf("unsupported backend type: %s", backendType)
	}

	return &Client{coreClient: coreClient}, nil
}

// SearchValues searches for unique values of a specified attribute
func (c *Client) SearchValues(backendName string, req backend.SearchValuesRequest) (backend.SearchValuesResponse, error) {
	return c.coreClient.SearchValues(backendName, req)
}

// TopTraces gets the top traces for a specified attribute and value
func (c *Client) TopTraces(backendName string, req backend.TopTracesRequest) (backend.TopTracesResponse, error) {
	return c.coreClient.TopTraces(backendName, req)
}

// ListSpans lists spans for a specified trace ID
func (c *Client) ListSpans(backendName string, req backend.ListSpansRequest) (backend.ListSpansResponse, error) {
	return c.coreClient.ListSpans(backendName, req)
}

// ListLogs lists logs for a specified trace ID
func (c *Client) ListLogs(backendName string, req backend.ListLogsRequest) (backend.ListLogsResponse, error) {
	return c.coreClient.ListLogs(backendName, req)
}

// GetAvailableBackends returns a list of available backends
func (c *Client) GetAvailableBackends() []string {
	return c.coreClient.GetAvailableBackends()
}

// Convenience functions that use the first available backend

// QuickSearchValues performs SearchValues using the first available backend
func (c *Client) QuickSearchValues(req backend.SearchValuesRequest) (backend.SearchValuesResponse, error) {
	backends := c.GetAvailableBackends()
	if len(backends) == 0 {
		return backend.SearchValuesResponse{}, fmt.Errorf("no backends available")
	}
	return c.SearchValues(backends[0], req)
}

// QuickTopTraces performs TopTraces using the first available backend
func (c *Client) QuickTopTraces(req backend.TopTracesRequest) (backend.TopTracesResponse, error) {
	backends := c.GetAvailableBackends()
	if len(backends) == 0 {
		return backend.TopTracesResponse{}, fmt.Errorf("no backends available")
	}
	return c.TopTraces(backends[0], req)
}

// QuickListSpans performs ListSpans using the first available backend
func (c *Client) QuickListSpans(req backend.ListSpansRequest) (backend.ListSpansResponse, error) {
	backends := c.GetAvailableBackends()
	if len(backends) == 0 {
		return backend.ListSpansResponse{}, fmt.Errorf("no backends available")
	}
	return c.ListSpans(backends[0], req)
}

// QuickListLogs performs ListLogs using the first available backend
func (c *Client) QuickListLogs(req backend.ListLogsRequest) (backend.ListLogsResponse, error) {
	backends := c.GetAvailableBackends()
	if len(backends) == 0 {
		return backend.ListLogsResponse{}, fmt.Errorf("no backends available")
	}
	return c.ListLogs(backends[0], req)
}
