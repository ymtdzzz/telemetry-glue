package pipeline

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ymtdzzz/telemetry-glue/pkg/analyzer"
	"github.com/ymtdzzz/telemetry-glue/pkg/output"
)

// PassthroughHandler handles stdin reading and data merging for pipeline commands
type PassthroughHandler struct {
	aggregator *analyzer.DataAggregator
}

// NewPassthroughHandler creates a new passthrough handler
func NewPassthroughHandler() *PassthroughHandler {
	return &PassthroughHandler{
		aggregator: analyzer.NewDataAggregator(),
	}
}

// ReadStdinIfAvailable reads from stdin if data is available and returns the aggregated data
func (p *PassthroughHandler) ReadStdinIfAvailable() (*analyzer.CombinedData, error) {
	// Check if stdin has data (non-interactive mode)
	stat, err := os.Stdin.Stat()
	if err != nil {
		return p.aggregator.GetCombinedData(), nil
	}

	// If stdin is a pipe or file, read from it
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		if err := p.aggregator.ReadFromStdin(os.Stdin); err != nil {
			// If we get a JSON parsing error but have no data, it might be empty stdin
			// In this case, just return empty data instead of failing
			if err.Error() == "failed to parse JSON: unexpected end of JSON input" {
				return p.aggregator.GetCombinedData(), nil
			}
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	return p.aggregator.GetCombinedData(), nil
}

// MergeSpansResult merges existing data with new spans result
func (p *PassthroughHandler) MergeSpansResult(existingData *analyzer.CombinedData, spansResult *output.SpansResult) *analyzer.CombinedData {
	merged := &analyzer.CombinedData{
		Spans:  existingData.Spans,
		Logs:   existingData.Logs,
		Traces: existingData.Traces,
		Values: existingData.Values,
	}

	// Add new spans
	merged.Spans = append(merged.Spans, spansResult.Spans...)

	return merged
}

// MergeLogsResult merges existing data with new logs result
func (p *PassthroughHandler) MergeLogsResult(existingData *analyzer.CombinedData, logsResult *output.LogsResult) *analyzer.CombinedData {
	merged := &analyzer.CombinedData{
		Spans:  existingData.Spans,
		Logs:   existingData.Logs,
		Traces: existingData.Traces,
		Values: existingData.Values,
	}

	// Add new logs
	merged.Logs = append(merged.Logs, logsResult.Logs...)

	return merged
}

// OutputMergedResult outputs the merged result as JSON
func (p *PassthroughHandler) OutputMergedResult(mergedData *analyzer.CombinedData, originalResult interface{}, format output.Format) error {
	// If no existing data, output original result
	if len(mergedData.Spans) == 0 && len(mergedData.Logs) == 0 && len(mergedData.Traces) == 0 && len(mergedData.Values) == 0 {
		// Get the original result interface and output it
		if formatter, ok := originalResult.(interface{ Print(output.Format) error }); ok {
			return formatter.Print(format)
		}
		return fmt.Errorf("original result does not support Print method")
	}

	// If we have merged data, output the combined result
	switch format {
	case output.FormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(mergedData)
	case output.FormatTable:
		// For table format, we'll output a summary
		fmt.Printf("Combined telemetry data:\n")
		fmt.Printf("- Spans: %d\n", len(mergedData.Spans))
		fmt.Printf("- Logs: %d\n", len(mergedData.Logs))
		fmt.Printf("- Traces: %d\n", len(mergedData.Traces))
		fmt.Printf("- Values: %d\n", len(mergedData.Values))
		return nil
	default:
		return fmt.Errorf("unsupported format for merged output: %s", format)
	}
}
