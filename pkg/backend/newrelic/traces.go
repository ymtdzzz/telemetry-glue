package newrelic

import (
	"fmt"
	"time"

	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	"github.com/ymtdzzz/telemetry-glue/pkg/output"
)

// TracesRequest represents a request to search for top traces in NewRelic
type TracesRequest struct {
	Entity    string // NewRelic entity name or GUID
	Attribute string // e.g. "http.path"
	Value     string // e.g. "/api/users" (exact match)
	Limit     int    // Number of top traces to return
	TimeRange TimeRange
}

// SearchTraces searches for top traces containing a specific attribute value
func (c *Client) SearchTraces(req TracesRequest) (*output.TracesResult, error) {
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
		return nil, fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	traces, err := c.parseTracesResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &output.TracesResult{
		Traces: traces,
	}, nil
}

// parseTracesResponse parses the NerdGraph response for SearchTraces
func (c *Client) parseTracesResponse(resp interface{}) ([]output.Trace, error) {
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

	var traces []output.Trace

	for _, result := range results {
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			continue
		}

		trace := output.Trace{}
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
