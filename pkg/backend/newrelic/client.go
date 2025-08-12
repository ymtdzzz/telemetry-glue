package newrelic

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
)

// init loads .env file if not in production environment
func init() {
	if os.Getenv("ENV") != "production" {
		_ = godotenv.Load()
	}
}

// TimeRange represents a time range for NewRelic queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Client represents a NewRelic client
type Client struct {
	client    nerdgraph.NerdGraph
	accountID int
}

// NewClient creates a new NewRelic client
func NewClient() (*Client, error) {
	// Get API key from environment variable
	apiKey := os.Getenv("NEW_RELIC_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	// Get account ID from environment variable
	accountIDStr := os.Getenv("NEW_RELIC_ACCOUNT_ID")
	if accountIDStr == "" {
		return nil, ErrMissingAccountID
	}

	accountID, err := strconv.Atoi(accountIDStr)
	if err != nil {
		return nil, ErrInvalidAccountID
	}

	// Initialize New Relic client
	cfg := config.New()
	cfg.PersonalAPIKey = apiKey
	client := nerdgraph.New(cfg)

	return &Client{
		client:    client,
		accountID: accountID,
	}, nil
}

// convertToString converts various types to string representation
func (c *Client) convertToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}
