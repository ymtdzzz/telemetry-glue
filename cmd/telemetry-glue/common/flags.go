package common

import (
	"fmt"
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
	Since  string
	Until  string
	Format string
}

// AddCommonFlags adds common flags to a command
func AddCommonFlags(cmd *cobra.Command, flags *CommonFlags) {
	cmd.Flags().StringVarP(&flags.Since, "since", "s", "1h", "Time range start (e.g., 1h, 30m, 24h)")
	cmd.Flags().StringVarP(&flags.Until, "until", "u", "", "Time range end (default: now)")
	cmd.Flags().StringVarP(&flags.Format, "format", "f", "table", "Output format (table, json, csv)")
}

// ParseTimeRange parses since/until duration strings into a TimeRange
func ParseTimeRange(since, until string) (TimeRange, error) {
	now := time.Now()

	// Parse since
	sinceDuration, err := time.ParseDuration(since)
	if err != nil {
		return TimeRange{}, fmt.Errorf("invalid since duration: %w", err)
	}
	startTime := now.Add(-sinceDuration)

	// Parse until (optional)
	var endTime time.Time
	if until != "" {
		untilDuration, err := time.ParseDuration(until)
		if err != nil {
			return TimeRange{}, fmt.Errorf("invalid until duration: %w", err)
		}
		endTime = now.Add(-untilDuration)
	} else {
		endTime = now
	}

	return TimeRange{
		Start: startTime,
		End:   endTime,
	}, nil
}

// ParseFormat parses format string and returns output.Format
func ParseFormat(formatStr string) (output.Format, error) {
	return output.ParseFormat(formatStr)
}
