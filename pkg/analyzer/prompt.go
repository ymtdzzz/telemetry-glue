package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
)

func generatePrompt(analysisType AnalysisType, telemetry *model.Telemetry, language string) (string, error) {
	switch analysisType {
	case AnalysisTypeDuration:
		return generateDurationPrompt(telemetry, language)
	case AnalysisTypeError:
		return generateErrorPrompt(telemetry, language)
	default:
		return "", fmt.Errorf("unsupported analysis type: %s", analysisType)
	}
}

// generateDurationPrompt generates a prompt for performance/duration analysis
func generateDurationPrompt(telemetry *model.Telemetry, language string) (string, error) {
	earliest, latest := telemetry.TimeRange()
	timeRange := ""
	if !earliest.IsZero() && !latest.IsZero() {
		timeRange = fmt.Sprintf("Time range: %s to %s (duration: %v)",
			earliest.Format(time.RFC3339),
			latest.Format(time.RFC3339),
			latest.Sub(earliest))
	}

	dataJSON, err := json.MarshalIndent(telemetry, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	prompt := fmt.Sprintf(`You are an expert in observability and performance analysis. Please analyze the following telemetry data for performance issues and bottlenecks.

## Data Summary
- Spans: %d entries
- Logs: %d entries  
%s

## Analysis Requirements
Please provide a comprehensive performance analysis including:

1. **Performance Bottlenecks**: Identify the slowest operations and services
2. **Duration Analysis**: Analyze span durations and identify outliers
3. **Critical Path**: Identify the critical path through the system
4. **Resource Utilization**: Look for signs of resource contention or inefficiency
5. **Correlation Analysis**: Correlate performance issues with logs and error patterns
6. **Optimization Recommendations**: Provide specific, actionable recommendations

## Output Format
Please structure your response as a markdown report with clear sections and bullet points.

## Telemetry Data
%s`,
		len(telemetry.Spans),
		len(telemetry.Logs),
		timeRange,
		string(dataJSON))

	// Add language-specific instructions
	if language == string(config.LanguageJapanese) {
		prompt += `

## Language Instructions
Please provide the analysis report in Japanese. Keep technical terms, metrics, and code snippets in English where appropriate. Structure the report with Japanese headers and explanations.`
	}

	return prompt, nil
}

// generateErrorPrompt generates a prompt for error analysis
func generateErrorPrompt(_ *model.Telemetry, _ string) (string, error) {
	return "", errors.New("error analysis prompt generation not implemented yet")
}
