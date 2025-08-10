package common

import (
	"strings"
	"testing"
	"time"
)

func TestParseAbsoluteTime(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "RFC3339 with Z",
			input:    "2024-01-15T10:00:00Z",
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "RFC3339 with timezone",
			input:    "2024-01-15T10:00:00+09:00",
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.FixedZone("", 9*60*60)),
			wantErr:  false,
		},
		{
			name:     "ISO8601 local time",
			input:    "2024-01-15T10:00:00",
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Date only",
			input:    "2024-01-15",
			expected: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Space separated with seconds",
			input:    "2024-01-15 10:00:00",
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:     "Space separated without seconds",
			input:    "2024-01-15 10:00",
			expected: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			wantErr:  false,
		},
		{
			name:    "Invalid format",
			input:   "invalid-time",
			wantErr: true,
		},
		{
			name:    "Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Invalid date",
			input:   "2024-13-50",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAbsoluteTime(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseAbsoluteTime() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("parseAbsoluteTime() unexpected error: %v", err)
				return
			}

			if !result.Equal(tt.expected) {
				t.Errorf("parseAbsoluteTime() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseTimeRange(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedStart time.Time
		expectedEnd   time.Time
		wantErr       bool
		errorContains string
	}{
		{
			name:          "Valid RFC3339 range",
			input:         "2024-01-15T10:00:00Z,2024-01-15T11:00:00Z",
			expectedStart: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
		{
			name:          "Valid date only range",
			input:         "2024-01-15,2024-01-16",
			expectedStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
		{
			name:          "Valid range with spaces",
			input:         " 2024-01-15T10:00:00Z , 2024-01-15T11:00:00Z ",
			expectedStart: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
		{
			name:          "Mixed formats",
			input:         "2024-01-15,2024-01-15T23:59:59Z",
			expectedStart: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC),
			wantErr:       false,
		},
		{
			name:    "Empty string returns default (last 1 hour)",
			input:   "",
			wantErr: false,
		},
		{
			name:          "Missing comma",
			input:         "2024-01-15T10:00:00Z 2024-01-15T11:00:00Z",
			wantErr:       true,
			errorContains: "time range must be in format 'from,to'",
		},
		{
			name:          "Too many parts",
			input:         "2024-01-15T10:00:00Z,2024-01-15T11:00:00Z,extra",
			wantErr:       true,
			errorContains: "time range must be in format 'from,to'",
		},
		{
			name:          "Invalid start time",
			input:         "invalid-start,2024-01-15T11:00:00Z",
			wantErr:       true,
			errorContains: "invalid start time",
		},
		{
			name:          "Invalid end time",
			input:         "2024-01-15T10:00:00Z,invalid-end",
			wantErr:       true,
			errorContains: "invalid end time",
		},
		{
			name:          "Start time after end time",
			input:         "2024-01-15T11:00:00Z,2024-01-15T10:00:00Z",
			wantErr:       true,
			errorContains: "start time",
		},
		{
			name:          "Same start and end time",
			input:         "2024-01-15T10:00:00Z,2024-01-15T10:00:00Z",
			expectedStart: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeRange(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseTimeRange() expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ParseTimeRange() error = %v, want error containing %v", err, tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseTimeRange() unexpected error: %v", err)
				return
			}

			// Special case for empty string (default behavior)
			if tt.input == "" {
				now := time.Now()
				expectedStart := now.Add(-time.Hour)

				// Allow some tolerance for timing differences in test execution
				if result.Start.Before(expectedStart.Add(-time.Second)) || result.Start.After(expectedStart.Add(time.Second)) {
					t.Errorf("ParseTimeRange() default start time not within expected range")
				}
				if result.End.Before(now.Add(-time.Second)) || result.End.After(now.Add(time.Second)) {
					t.Errorf("ParseTimeRange() default end time not within expected range")
				}
				return
			}

			if !result.Start.Equal(tt.expectedStart) {
				t.Errorf("ParseTimeRange() start = %v, want %v", result.Start, tt.expectedStart)
			}
			if !result.End.Equal(tt.expectedEnd) {
				t.Errorf("ParseTimeRange() end = %v, want %v", result.End, tt.expectedEnd)
			}
		})
	}
}
