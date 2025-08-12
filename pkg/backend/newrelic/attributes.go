package newrelic

import (
	"fmt"
	"strings"
	"time"

	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/ymtdzzz/telemetry-glue/pkg/output"
)

// AttributesRequest represents a request to search for attribute values in NewRelic
type AttributesRequest struct {
	Entity    string // NewRelic entity name or GUID
	Attribute string // e.g. "http.path"
	Query     string // e.g. "*user*"
	TimeRange TimeRange
}

// Attributes searches for unique values of a specified attribute
func (c *Client) Attributes(req AttributesRequest) (*output.AttributesResult, error) {
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
		return nil, fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	values, err := c.parseAttributesResponse(resp, req.Attribute)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &output.AttributesResult{
		Values: values,
	}, nil
}

// parseAttributesResponse parses the NerdGraph response for string arrays
func (c *Client) parseAttributesResponse(resp interface{}, attribute string) ([]string, error) {
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
