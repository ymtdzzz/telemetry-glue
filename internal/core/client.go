// Package core provides the core query integration layer for telemetry-glue
package core

import (
	"fmt"

	"github.com/ymtdzzz/telemetry-glue/internal/backend"
)

// Client represents the core telemetry-glue client that integrates multiple backends
type Client struct {
	backends map[string]backend.Backend
}

// NewClient creates a new core client instance
func NewClient() (*Client, error) {
	return &Client{
		backends: make(map[string]backend.Backend),
	}, nil
}

// AddBackend adds a backend to the client
func (c *Client) AddBackend(name string, b backend.Backend) {
	c.backends[name] = b
}

// GetBackend returns a backend by name
func (c *Client) GetBackend(name string) (backend.Backend, error) {
	b, exists := c.backends[name]
	if !exists {
		return nil, fmt.Errorf("backend not found: %s", name)
	}
	return b, nil
}

// SearchValues searches for unique values of a specified attribute using the specified backend
func (c *Client) SearchValues(backendName string, req backend.SearchValuesRequest) (backend.SearchValuesResponse, error) {
	b, err := c.GetBackend(backendName)
	if err != nil {
		return backend.SearchValuesResponse{}, err
	}

	return b.SearchValues(req)
}

// TopTraces gets the top traces for a specified attribute and value using the specified backend
func (c *Client) TopTraces(backendName string, req backend.TopTracesRequest) (backend.TopTracesResponse, error) {
	b, err := c.GetBackend(backendName)
	if err != nil {
		return backend.TopTracesResponse{}, err
	}

	return b.TopTraces(req)
}

// ListSpans lists spans for a specified trace ID using the specified backend
func (c *Client) ListSpans(backendName string, req backend.ListSpansRequest) (backend.ListSpansResponse, error) {
	b, err := c.GetBackend(backendName)
	if err != nil {
		return backend.ListSpansResponse{}, err
	}

	return b.ListSpans(req)
}

// ListLogs lists logs for a specified trace ID using the specified backend
func (c *Client) ListLogs(backendName string, req backend.ListLogsRequest) (backend.ListLogsResponse, error) {
	b, err := c.GetBackend(backendName)
	if err != nil {
		return backend.ListLogsResponse{}, err
	}

	return b.ListLogs(req)
}

// GetAvailableBackends returns a list of initialized backends
func (c *Client) GetAvailableBackends() []string {
	backends := make([]string, 0, len(c.backends))
	for name := range c.backends {
		backends = append(backends, name)
	}
	return backends
}
