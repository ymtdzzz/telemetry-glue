package newrelic

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
)

// Common errors for NewRelic backend
var (
	ErrMissingAPIKey    = errors.New("NEW_RELIC_API_KEY is required")
	ErrMissingAccountID = errors.New("NEW_RELIC_ACCOUNT_ID is required")
	ErrInvalidAccountID = errors.New("invalid NEW_RELIC_ACCOUNT_ID format")
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

// SearchValuesRequest represents a request to search for attribute values in NewRelic
type SearchValuesRequest struct {
	Entity    string // NewRelic entity name or GUID
	Attribute string // e.g. "http.path"
	Query     string // e.g. "*user*"
	TimeRange TimeRange
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

// SearchValues searches for unique values of a specified attribute
func (c *Client) SearchValues(req SearchValuesRequest) ([]string, string, error) {
	// Build NRQL query for searching attribute values with pattern matching
	// Convert wildcard pattern (*user*) to SQL LIKE pattern (%user%)
	likePattern := strings.ReplaceAll(req.Query, "*", "%")

	// Calculate time range in minutes from current time
	timeSinceStart := time.Since(req.TimeRange.Start).Minutes()

	nrqlQuery := fmt.Sprintf(
		"SELECT uniques(%s) FROM Span WHERE %s LIKE '%s' SINCE %d minutes ago UNTIL %d minutes ago",
		req.Attribute,
		req.Attribute,
		likePattern,
		int(timeSinceStart),
		int(time.Since(req.TimeRange.End).Minutes()),
	)

	// Build GraphQL query
	graphqlQuery := `
		query($accountId: Int!, $nrqlQuery: Nrql!) {
			actor {
				account(id: $accountId) {
					nrql(query: $nrqlQuery, timeout: 30) {
						results
					}
				}
			}
		}`

	variables := map[string]interface{}{
		"accountId": c.accountID,
		"nrqlQuery": nrqlQuery,
	}

	// Execute the query
	resp, err := c.client.Query(graphqlQuery, variables)
	if err != nil {
		return nil, "", fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	values, err := c.parseSearchValuesResponse(resp, req.Attribute)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate web link
	webLink := c.generateWebLinkForSearchValues(req.Attribute, req.Query, req.TimeRange)

	return values, webLink, nil
}

// parseSearchValuesResponse parses the NerdGraph response for SearchValues
func (c *Client) parseSearchValuesResponse(resp interface{}, attribute string) ([]string, error) {
	// First, assert the response as QueryResponse type
	queryResp, ok := resp.(nerdgraph.QueryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	// Parse the Actor field as map[string]interface{}
	actor, ok := queryResp.Actor.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("actor not found in response")
	}

	account, ok := actor["account"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("account not found in response")
	}

	nrql, ok := account["nrql"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("nrql not found in response")
	}

	results, ok := nrql["results"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("results not found in response")
	}

	var values []string
	uniqueKey := fmt.Sprintf("uniques.%s", attribute)

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		if uniqueValues, exists := resultMap[uniqueKey]; exists {
			if valuesList, ok := uniqueValues.([]interface{}); ok {
				for _, val := range valuesList {
					if strVal, err := c.convertToString(val); err == nil {
						values = append(values, strVal)
					}
				}
			}
		}
	}

	return values, nil
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

// generateWebLinkForSearchValues generates a New Relic UI link for search values
func (c *Client) generateWebLinkForSearchValues(attribute, query string, timeRange TimeRange) string {
	// Generate NRQL query for the web link
	nrqlQuery := fmt.Sprintf("SELECT %s FROM Span WHERE %s LIKE '%%%s%%' SINCE %d minutes ago",
		attribute, attribute, query, int(time.Since(timeRange.Start).Minutes()))

	// New Relic query link format
	return fmt.Sprintf("https://one.newrelic.com/nr1-core?account=%d&filters=%%7B%%22query%%22%%3A%%22%s%%22%%7D",
		c.accountID, nrqlQuery)
}
