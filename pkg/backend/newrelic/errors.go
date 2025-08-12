package newrelic

import "errors"

// Common errors for NewRelic backend
var (
	ErrMissingAPIKey    = errors.New("NEW_RELIC_API_KEY is required")
	ErrMissingAccountID = errors.New("NEW_RELIC_ACCOUNT_ID is required")
	ErrInvalidAccountID = errors.New("invalid NEW_RELIC_ACCOUNT_ID format")
)
