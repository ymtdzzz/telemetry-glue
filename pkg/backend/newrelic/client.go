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

// TopTracesRequest represents a request to search for top traces in NewRelic
type TopTracesRequest struct {
	Entity    string // NewRelic entity name or GUID
	Attribute string // e.g. "http.path"
	Value     string // e.g. "/api/users" (exact match)
	Limit     int    // Number of top traces to return
	TimeRange TimeRange
}

// SpansRequest represents a request to search for spans in NewRelic
type SpansRequest struct {
	TraceID   string // Trace ID to get spans from
	TimeRange TimeRange
}

// TraceInfo represents information about a trace
type TraceInfo struct {
	TraceID     string
	Duration    float64 // Duration in milliseconds
	ServiceName string
	SpanCount   int
	StartTime   time.Time
}

// SpanInfo represents information about a span
type SpanInfo map[string]interface{}

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

// SearchTopTraces searches for top traces containing a specific attribute value
func (c *Client) SearchTopTraces(req TopTracesRequest) ([]TraceInfo, string, error) {
	// Calculate time range in minutes from current time
	timeSinceStart := time.Since(req.TimeRange.Start).Minutes()
	timeSinceEnd := time.Since(req.TimeRange.End).Minutes()

	// Build NRQL query to find top traces by duration
	// First find traces that contain spans with the specified attribute value
	nrqlQuery := fmt.Sprintf(`
		SELECT 
			max(duration.ms) as maxDuration,
			earliest(service.name) as serviceName,
			count(*) as spanCount,
			earliest(timestamp) as startTime
		FROM Span 
		WHERE %s = '%s' 
		SINCE %d minutes ago UNTIL %d minutes ago 
		FACET trace.id 
		ORDER BY maxDuration DESC 
		LIMIT %d`,
		req.Attribute,
		req.Value,
		int(timeSinceStart),
		int(timeSinceEnd),
		req.Limit,
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
	traces, err := c.parseTopTracesResponse(resp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate web link
	webLink := c.generateWebLinkForTopTraces(req.Attribute, req.Value, req.TimeRange)

	return traces, webLink, nil
}

// SearchSpans searches for all spans within a specific trace
func (c *Client) SearchSpans(req SpansRequest) ([]SpanInfo, string, error) {
	// Calculate time range in minutes from current time
	timeSinceStart := time.Since(req.TimeRange.Start).Minutes()
	timeSinceEnd := time.Since(req.TimeRange.End).Minutes()

	// Build NRQL query to get all spans for the trace
	nrqlQuery := fmt.Sprintf(`
		SELECT * 
		FROM Span 
		WHERE trace.id = '%s' 
		SINCE %d minutes ago UNTIL %d minutes ago 
		ORDER BY timestamp ASC`,
		req.TraceID,
		int(timeSinceStart),
		int(timeSinceEnd),
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
	spans, err := c.parseSpansResponse(resp)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate web link
	webLink := c.generateWebLinkForSpans(req.TraceID, req.TimeRange)

	return spans, webLink, nil
}

// parseSpansResponse parses the NerdGraph response for SearchSpans
func (c *Client) parseSpansResponse(resp interface{}) ([]SpanInfo, error) {
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

	var spans []SpanInfo

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		// Create SpanInfo as a map containing all the data from NewRelic
		span := make(SpanInfo)
		for key, value := range resultMap {
			span[key] = value
		}

		spans = append(spans, span)
	}

	return spans, nil
}

// generateWebLinkForSpans generates a New Relic UI link for spans
func (c *Client) generateWebLinkForSpans(traceID string, timeRange TimeRange) string {
	// Generate NRQL query for the web link
	nrqlQuery := fmt.Sprintf(`SELECT * FROM Span WHERE trace.id = '%s' SINCE %d minutes ago ORDER BY timestamp ASC`,
		traceID, int(time.Since(timeRange.Start).Minutes()))

	// New Relic query link format
	return fmt.Sprintf("https://one.newrelic.com/nr1-core?account=%d&filters=%%7B%%22query%%22%%3A%%22%s%%22%%7D",
		c.accountID, nrqlQuery)
}

// parseTopTracesResponse parses the NerdGraph response for SearchTopTraces
func (c *Client) parseTopTracesResponse(resp interface{}) ([]TraceInfo, error) {
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

	var traces []TraceInfo

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		trace := TraceInfo{}
		// Extract trace.id from facet (when using FACET, the key is the faceted value)
		// Try multiple possible locations for the trace ID
		if facetArray, exists := resultMap["facet"]; exists {
			if facetSlice, ok := facetArray.([]interface{}); ok && len(facetSlice) > 0 {
				if traceIDStr, err := c.convertToString(facetSlice[0]); err == nil {
					trace.TraceID = traceIDStr
				}
			}
		}

		// Also try looking for trace.id directly in the result
		if traceID, exists := resultMap["trace.id"]; exists {
			if traceIDStr, err := c.convertToString(traceID); err == nil {
				trace.TraceID = traceIDStr
			}
		}

		// If trace ID is still empty, check for other possible keys
		if trace.TraceID == "" {
			// Sometimes it might be under a different key, let's try to find any key that looks like a trace ID
			for key, value := range resultMap {
				if key == "trace.id" || key == "traceId" {
					if traceIDStr, err := c.convertToString(value); err == nil {
						trace.TraceID = traceIDStr
						break
					}
				}
			}
		}

		// Extract maxDuration
		if duration, exists := resultMap["maxDuration"]; exists {
			if durationFloat, ok := duration.(float64); ok {
				trace.Duration = durationFloat
			}
		}

		// Extract serviceName
		if serviceName, exists := resultMap["serviceName"]; exists {
			if serviceNameStr, err := c.convertToString(serviceName); err == nil {
				trace.ServiceName = serviceNameStr
			}
		}

		// Extract spanCount
		if spanCount, exists := resultMap["spanCount"]; exists {
			if spanCountFloat, ok := spanCount.(float64); ok {
				trace.SpanCount = int(spanCountFloat)
			}
		}

		// Extract startTime
		if startTime, exists := resultMap["startTime"]; exists {
			if startTimeFloat, ok := startTime.(float64); ok {
				// Convert from Unix timestamp in milliseconds
				trace.StartTime = time.Unix(int64(startTimeFloat/1000), 0)
			}
		}

		traces = append(traces, trace)
	}

	return traces, nil
}

// generateWebLinkForTopTraces generates a New Relic UI link for top traces
func (c *Client) generateWebLinkForTopTraces(attribute, value string, timeRange TimeRange) string {
	// Generate NRQL query for the web link
	nrqlQuery := fmt.Sprintf(`SELECT max(duration.ms) as maxDuration FROM Span WHERE %s = '%s' SINCE %d minutes ago FACET trace.id ORDER BY maxDuration DESC`,
		attribute, value, int(time.Since(timeRange.Start).Minutes()))

	// New Relic query link format
	return fmt.Sprintf("https://one.newrelic.com/nr1-core?account=%d&filters=%%7B%%22query%%22%%3A%%22%s%%22%%7D",
		c.accountID, nrqlQuery)
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
