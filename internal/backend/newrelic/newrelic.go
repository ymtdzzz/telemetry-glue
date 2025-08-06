package newrelic

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/ymtdzzz/telemetry-glue/internal/backend"
)

// init loads .env file if not in production environment
func init() {
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}
}

// NewRelicBackend implements the Backend interface for New Relic
type NewRelicBackend struct {
	client    nerdgraph.NerdGraph
	accountID int
}

// NewNewRelicBackend creates a new NewRelic backend instance
func NewNewRelicBackend() (*NewRelicBackend, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("NEW_RELIC_API_KEY")
	if apiKey == "" {
		return nil, backend.ErrMissingAPIKey
	}

	// Get account ID from environment variable
	accountIDStr := os.Getenv("NEW_RELIC_ACCOUNT_ID")
	if accountIDStr == "" {
		return nil, backend.ErrMissingAccountID
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		return nil, backend.ErrInvalidAccountID
	}

	// Initialize New Relic client
	cfg := config.New()
	cfg.PersonalAPIKey = apiKey
	client := nerdgraph.New(cfg)

	return &NewRelicBackend{
		client:    client,
		accountID: accountID,
	}, nil
}

// Name returns the name of this backend
func (nr *NewRelicBackend) Name() string {
	return "newrelic"
}

// SearchValues implements the SearchValues method of the Backend interface
func (nr *NewRelicBackend) SearchValues(req backend.SearchValuesRequest) (backend.SearchValuesResponse, error) {
	return nr.searchValuesImpl(req)
}

// TopTraces implements the TopTraces method of the Backend interface
func (nr *NewRelicBackend) TopTraces(req backend.TopTracesRequest) (backend.TopTracesResponse, error) {
	// TODO: Implement TopTraces
	return backend.TopTracesResponse{}, nil
}

// ListSpans implements the ListSpans method of the Backend interface
func (nr *NewRelicBackend) ListSpans(req backend.ListSpansRequest) (backend.ListSpansResponse, error) {
	// TODO: Implement ListSpans
	return backend.ListSpansResponse{}, nil
}

// ListLogs implements the ListLogs method of the Backend interface
func (nr *NewRelicBackend) ListLogs(req backend.ListLogsRequest) (backend.ListLogsResponse, error) {
	// TODO: Implement ListLogs
	return backend.ListLogsResponse{}, nil
}
