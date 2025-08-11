package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// AnalysisType represents the type of analysis to perform
type AnalysisType string

const (
	AnalysisTypeDuration AnalysisType = "duration"
	AnalysisTypeError    AnalysisType = "error"
)

// PromptGenerator generates prompts for different analysis types
type PromptGenerator struct{}

// NewPromptGenerator creates a new prompt generator
func NewPromptGenerator() *PromptGenerator {
	return &PromptGenerator{}
}

// GeneratePrompt generates a prompt based on analysis type and combined data
func (pg *PromptGenerator) GeneratePrompt(analysisType AnalysisType, data *CombinedData, language string) (string, error) {
	switch analysisType {
	case AnalysisTypeDuration:
		return pg.generateDurationPrompt(data, language)
	case AnalysisTypeError:
		return pg.generateErrorPrompt(data, language)
	default:
		return "", fmt.Errorf("unsupported analysis type: %s", analysisType)
	}
}

// generateDurationPrompt generates a prompt for performance/duration analysis
func (pg *PromptGenerator) generateDurationPrompt(data *CombinedData, language string) (string, error) {
	earliest, latest := data.GetTimeRange()
	timeRange := ""
	if !earliest.IsZero() && !latest.IsZero() {
		timeRange = fmt.Sprintf("Time range: %s to %s (duration: %v)",
			earliest.Format(time.RFC3339),
			latest.Format(time.RFC3339),
			latest.Sub(earliest))
	}

	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	prompt := fmt.Sprintf(`You are an expert in observability and performance analysis. Please analyze the following telemetry data for performance issues and bottlenecks.

## Data Summary
- Spans: %d entries
- Logs: %d entries  
- Traces: %d entries
- Values: %d entries
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
		len(data.Spans),
		len(data.Logs),
		len(data.Traces),
		len(data.Values),
		timeRange,
		string(dataJSON))

	// Add language-specific instructions
	if language == "ja" {
		prompt += `

## Language Instructions
Please provide the analysis report in Japanese. Keep technical terms, metrics, and code snippets in English where appropriate. Structure the report with Japanese headers and explanations.`
	}

	return prompt, nil
}

// generateErrorPrompt generates a prompt for error analysis
func (pg *PromptGenerator) generateErrorPrompt(data *CombinedData, language string) (string, error) {
	earliest, latest := data.GetTimeRange()
	timeRange := ""
	if !earliest.IsZero() && !latest.IsZero() {
		timeRange = fmt.Sprintf("Time range: %s to %s (duration: %v)",
			earliest.Format(time.RFC3339),
			latest.Format(time.RFC3339),
			latest.Sub(earliest))
	}

	// Count error indicators
	errorSpans := 0
	errorLogs := 0

	for _, span := range data.Spans {
		if pg.isErrorSpan(span) {
			errorSpans++
		}
	}

	for _, log := range data.Logs {
		if pg.isErrorLog(log.Message) {
			errorLogs++
		}
	}

	dataJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	prompt := fmt.Sprintf(`You are an expert in error analysis and system reliability. Please analyze the following telemetry data to identify error patterns and root causes.

## Data Summary
- Spans: %d entries (%d with errors)
- Logs: %d entries (%d with error indicators)
- Traces: %d entries
- Values: %d entries
%s

## Analysis Requirements
Please provide a comprehensive error analysis including:

1. **Error Patterns**: Identify common error types and patterns
2. **Root Cause Analysis**: Analyze the chain of events leading to errors
3. **Error Distribution**: Show how errors are distributed across services/time
4. **Impact Assessment**: Assess the business impact of identified errors
5. **Correlation Analysis**: Correlate errors between spans and logs
6. **Recovery Patterns**: Identify if there are retry patterns or recovery mechanisms
7. **Remediation Steps**: Provide specific steps to resolve identified issues

## Output Format
Please structure your response as a markdown report with clear sections and bullet points.

## Telemetry Data
%s`,
		len(data.Spans),
		errorSpans,
		len(data.Logs),
		errorLogs,
		len(data.Traces),
		len(data.Values),
		timeRange,
		string(dataJSON))

	// Add language-specific instructions
	if language == "ja" {
		prompt += `

## Language Instructions
Please provide the analysis report in Japanese. Keep technical terms, metrics, and code snippets in English where appropriate. Structure the report with Japanese headers and explanations.`
	}

	return prompt, nil
}

// isErrorSpan checks if a span indicates an error
func (pg *PromptGenerator) isErrorSpan(span map[string]interface{}) bool {
	// Check for error tags or status codes
	if status, ok := span["http.status_code"].(float64); ok && status >= 400 {
		return true
	}

	if error, ok := span["error"].(bool); ok && error {
		return true
	}

	if errorMsg, ok := span["error.message"].(string); ok && errorMsg != "" {
		return true
	}

	return false
}

// isErrorLog checks if a log message indicates an error
func (pg *PromptGenerator) isErrorLog(message string) bool {
	errorKeywords := []string{
		"error", "ERROR", "Error",
		"exception", "EXCEPTION", "Exception",
		"fail", "FAIL", "Fail",
		"fatal", "FATAL", "Fatal",
		"panic", "PANIC", "Panic",
	}

	lowerMessage := strings.ToLower(message)
	for _, keyword := range errorKeywords {
		if strings.Contains(lowerMessage, strings.ToLower(keyword)) {
			return true
		}
	}

	return false
}
