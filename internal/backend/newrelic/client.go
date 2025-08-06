package newrelic

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/ymtdzzz/telemetry-glue/internal/backend"
)

// searchValuesImpl implements the core SearchValues functionality
func (nr *NewRelicBackend) searchValuesImpl(req backend.SearchValuesRequest) (backend.SearchValuesResponse, error) {
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
		"accountId": nr.accountID,
		"nrqlQuery": nrqlQuery,
	}

	// Execute the query
	resp, err := nr.client.Query(graphqlQuery, variables)
	if err != nil {
		return backend.SearchValuesResponse{}, fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	values, err := nr.parseSearchValuesResponse(resp, req.Attribute)
	if err != nil {
		return backend.SearchValuesResponse{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Generate web link
	webLink := generateWebLinkForSearchValues(nr.accountID, req.Attribute, req.Query, req.TimeRange)

	return backend.SearchValuesResponse{
		Values:  values,
		WebLink: webLink,
	}, nil
}

// parseSearchValuesResponse parses the NerdGraph response for SearchValues
func (nr *NewRelicBackend) parseSearchValuesResponse(resp interface{}, attribute string) ([]string, error) {
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
					if strVal, err := nr.convertToString(val); err == nil {
						values = append(values, strVal)
					}
				}
			}
		}
	}

	return values, nil
}

// convertToString converts various types to string representation
func (nr *NewRelicBackend) convertToString(value interface{}) (string, error) {
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
