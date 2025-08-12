package newrelic

import (
	"fmt"
	"time"

	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/ymtdzzz/telemetry-glue/pkg/output"
)

// SpansRequest represents a request to search for spans in NewRelic
type SpansRequest struct {
	TraceID   string // Trace ID to get spans from
	TimeRange TimeRange
}

// SearchSpans searches for all spans within a specific trace
func (c *Client) SearchSpans(req SpansRequest) (*output.SpansResult, error) {
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
		return nil, fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	spans, err := c.parseSpansResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &output.SpansResult{
		Spans: spans,
	}, nil
}

// parseSpansResponse parses the NerdGraph response for SearchSpans
func (c *Client) parseSpansResponse(resp interface{}) ([]output.Span, error) {
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

	var spans []output.Span

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		// Create SpanInfo as a map containing all the data from NewRelic
		span := make(output.Span)
		for key, value := range resultMap {
			span[key] = value
		}

		spans = append(spans, span)
	}

	return spans, nil
}
