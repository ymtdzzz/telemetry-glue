package common

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ymtdzzz/telemetry-glue/internal/output"
)

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// CommonFlags holds common CLI flags used across commands
type CommonFlags struct {
	TimeRange string
	Format    string
}

// AddCommonFlags adds common flags to a command
func AddCommonFlags(cmd *cobra.Command, flags *CommonFlags) {
	cmd.Flags().StringVarP(&flags.TimeRange, "time-range", "t", "", "Time range in format 'from,to' (e.g., '2024-01-15T10:00:00Z,2024-01-15T11:00:00Z'). Default: last 1 hour")
	cmd.Flags().StringVarP(&flags.Format, "format", "f", "table", "Output format (table, json, csv)")
}

// ParseTimeRange parses comma-separated absolute time strings into a TimeRange
func ParseTimeRange(timeRange string) (TimeRange, error) {
	now := time.Now()

	// If empty, default to last 1 hour
	if timeRange == "" {
		return TimeRange{
			Start: now.Add(-time.Hour),
			End:   now,
		}, nil
	}

	// Split by comma
	parts := strings.Split(timeRange, ",")
	if len(parts) != 2 {
		return TimeRange{}, fmt.Errorf("time range must be in format 'from,to', got: %s", timeRange)
	}

	fromStr := strings.TrimSpace(parts[0])
	toStr := strings.TrimSpace(parts[1])

	// Parse start time
	startTime, err := parseAbsoluteTime(fromStr)
	if err != nil {
		return TimeRange{}, fmt.Errorf("invalid start time '%s': %w", fromStr, err)
	}

	// Parse end time
	endTime, err := parseAbsoluteTime(toStr)
	if err != nil {
		return TimeRange{}, fmt.Errorf("invalid end time '%s': %w", toStr, err)
	}

	// Validate that start is before end
	if startTime.After(endTime) {
		return TimeRange{}, fmt.Errorf("start time (%s) must be before end time (%s)", startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	}

	return TimeRange{
		Start: startTime,
		End:   endTime,
	}, nil
}

// parseAbsoluteTime parses various absolute time formats
func parseAbsoluteTime(timeStr string) (time.Time, error) {
	// Try different time formats in order of preference
	formats := []string{
		time.RFC3339,          // 2024-01-15T10:00:00Z or 2024-01-15T10:00:00+09:00
		"2006-01-02T15:04:05", // 2024-01-15T10:00:00 (local time)
		"2006-01-02",          // 2024-01-15 (date only, 00:00:00)
		"2006-01-02 15:04:05", // 2024-01-15 10:00:00 (space separated)
		"2006-01-02 15:04",    // 2024-01-15 10:00 (without seconds)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time '%s'. Supported formats: RFC3339 (2024-01-15T10:00:00Z), ISO8601 (2024-01-15T10:00:00), date only (2024-01-15)", timeStr)
}

// ParseFormat parses format string and returns output.Format
func ParseFormat(formatStr string) (output.Format, error) {
	return output.ParseFormat(formatStr)
}
