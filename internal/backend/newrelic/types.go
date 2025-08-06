package newrelic

import (
	"fmt"
	"time"

	"github.com/ymtdzzz/telemetry-glue/internal/backend"
)

// generateWebLinkForSearchValues generates a New Relic UI link for search values
func generateWebLinkForSearchValues(accountID int, attribute, query string, timeRange backend.TimeRange) string {
	// Generate NRQL query for the web link
	nrqlQuery := fmt.Sprintf("SELECT %s FROM Span WHERE %s LIKE '%%%s%%' SINCE %d minutes ago",
		attribute, attribute, query, int(time.Since(timeRange.Start).Minutes()))

	// New Relic query link format
	return fmt.Sprintf("https://one.newrelic.com/nr1-core?account=%d&filters=%%7B%%22query%%22%%3A%%22%s%%22%%7D",
		accountID, nrqlQuery)
}
