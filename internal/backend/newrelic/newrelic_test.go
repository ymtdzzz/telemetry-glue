package newrelic

import (
	"testing"
	"time"

	"github.com/ymtdzzz/telemetry-glue/internal/backend"
)

func TestNewNewRelicBackend(t *testing.T) {
	// Set environment variables for testing
	t.Setenv("NEW_RELIC_API_KEY", "test-api-key")
	t.Setenv("NEW_RELIC_ACCOUNT_ID", "12345")

	backend, err := NewNewRelicBackend()
	if err != nil {
		t.Fatalf("Failed to create NewRelic backend: %v", err)
	}

	if backend.Name() != "newrelic" {
		t.Errorf("Expected backend name 'newrelic', got '%s'", backend.Name())
	}

	if backend.accountID != 12345 {
		t.Errorf("Expected account ID 12345, got %d", backend.accountID)
	}
}

func TestNewNewRelicBackend_MissingAPIKey(t *testing.T) {
	t.Setenv("NEW_RELIC_API_KEY", "")
	t.Setenv("NEW_RELIC_ACCOUNT_ID", "12345")

	_, err := NewNewRelicBackend()
	if err != backend.ErrMissingAPIKey {
		t.Errorf("Expected ErrMissingAPIKey, got %v", err)
	}
}

func TestNewNewRelicBackend_MissingAccountID(t *testing.T) {
	t.Setenv("NEW_RELIC_API_KEY", "test-api-key")
	t.Setenv("NEW_RELIC_ACCOUNT_ID", "")

	_, err := NewNewRelicBackend()
	if err != backend.ErrMissingAccountID {
		t.Errorf("Expected ErrMissingAccountID, got %v", err)
	}
}

func TestNewNewRelicBackend_InvalidAccountID(t *testing.T) {
	t.Setenv("NEW_RELIC_API_KEY", "test-api-key")
	t.Setenv("NEW_RELIC_ACCOUNT_ID", "invalid")

	_, err := NewNewRelicBackend()
	if err != backend.ErrInvalidAccountID {
		t.Errorf("Expected ErrInvalidAccountID, got %v", err)
	}
}

func TestParseSearchValuesResponse(t *testing.T) {
	// Create a mock backend for testing parsing logic
	mockBackend := &NewRelicBackend{accountID: 12345}

	// Test case 1: Valid response with uniques results
	validResponse := map[string]interface{}{
		"actor": map[string]interface{}{
			"account": map[string]interface{}{
				"nrql": map[string]interface{}{
					"results": []interface{}{
						map[string]interface{}{
							"uniques.http.path": []interface{}{
								"/admin/users/new",
								"/user/profile",
								"/api/users/create",
							},
						},
					},
				},
			},
		},
	}

	values, err := mockBackend.parseSearchValuesResponse(validResponse, "http.path")
	if err != nil {
		t.Fatalf("Failed to parse valid response: %v", err)
	}

	expected := []string{"/admin/users/new", "/user/profile", "/api/users/create"}
	if len(values) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, expectedValue := range expected {
		if i >= len(values) || values[i] != expectedValue {
			t.Errorf("Expected value '%s' at index %d, got '%s'", expectedValue, i, values[i])
		}
	}
}

func TestConvertToString(t *testing.T) {
	mockBackend := &NewRelicBackend{}

	testCases := []struct {
		input    interface{}
		expected string
	}{
		{"string_value", "string_value"},
		{42.5, "42.5"},
		{123, "123"},
		{int64(456), "456"},
		{true, "true"},
		{false, "false"},
	}

	for _, tc := range testCases {
		result, err := mockBackend.convertToString(tc.input)
		if err != nil {
			t.Errorf("Failed to convert %v: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("Expected '%s', got '%s' for input %v", tc.expected, result, tc.input)
		}
	}
}

func TestSearchValuesRequest_TimeRange(t *testing.T) {
	now := time.Now()
	oneHourAgo := now.Add(-time.Hour)

	req := backend.SearchValuesRequest{
		Attribute: "http.path",
		Query:     "*user*",
		TimeRange: backend.TimeRange{
			Start: oneHourAgo,
			End:   now,
		},
	}

	// Verify that TimeRange is properly structured
	if req.TimeRange.Start.After(req.TimeRange.End) {
		t.Error("Start time should be before end time")
	}

	duration := req.TimeRange.End.Sub(req.TimeRange.Start)
	if duration != time.Hour {
		t.Errorf("Expected duration of 1 hour, got %v", duration)
	}
}
