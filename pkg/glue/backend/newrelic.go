package backend

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maps"

	"github.com/newrelic/newrelic-client-go/v2/pkg/config"
	"github.com/newrelic/newrelic-client-go/v2/pkg/nerdgraph"
	gconfig "github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
)

// NewRelicBackend represents a NewRelic backend
type NewRelicBackend struct {
	client    *nerdgraph.NerdGraph
	accountID int
}

// NewNewRelicBackend creates a new NewRelic backend
func NewNewRelicBackend(cfg *gconfig.NewRelicConfig) *NewRelicBackend {
	nrcfg := config.New()
	nrcfg.PersonalAPIKey = cfg.APIKey
	client := nerdgraph.New(nrcfg)

	return &NewRelicBackend{
		client:    &client,
		accountID: cfg.AccountID,
	}
}

func (n *NewRelicBackend) SearchSpans(ctx context.Context, req *SearchSpansRequest) (model.Spans, error) {
	// Build NRQL query to get all spans for the trace
	nrqlQuery := fmt.Sprintf(`
		SELECT * 
		FROM Span 
		WHERE trace.id = '%s' 
		SINCE %d UNTIL %d 
		ORDER BY timestamp ASC`,
		req.TraceID,
		req.TimeRange.Start.UnixMilli(),
		req.TimeRange.End.UnixMilli(),
	)

	log.Printf("Executing NRQL query: %s", nrqlQuery)

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

	variables := map[string]any{
		"accountId": n.accountID,
		"nrqlQuery": nrqlQuery,
	}

	// Execute the query
	resp, err := n.client.Query(graphqlQuery, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to execute NerdGraph query: %w", err)
	}

	// Parse the response
	spans, err := n.parseSpansResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return spans, nil
}

// parseSpansResponse parses the NerdGraph response for SearchSpans
func (n *NewRelicBackend) parseSpansResponse(resp any) (model.Spans, error) {
	// First, assert the response as QueryResponse type
	queryResp, ok := resp.(nerdgraph.QueryResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", resp)
	}

	// Parse the Actor field as map[string]any
	actor, ok := queryResp.Actor.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("actor not found in response")
	}

	account, ok := actor["account"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("account not found in response")
	}

	nrql, ok := account["nrql"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("nrql not found in response")
	}

	results, ok := nrql["results"].([]any)
	if !ok {
		return nil, fmt.Errorf("results not found in response")
	}

	var spans model.Spans

	for _, result := range results {
		resultMap, ok := result.(map[string]any)
		if !ok {
			continue
		}

		span := make(model.Span)
		maps.Copy(span, resultMap)

		spans = append(spans, span)
	}

	return spans, nil
}

func (n *NewRelicBackend) SearchLogs(ctx context.Context, req *SearchLogsRequest) (model.Logs, error) {
	return nil, errors.New("not implemented")
}
