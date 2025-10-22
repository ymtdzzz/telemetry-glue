package analyzer

import (
	"context"

	"github.com/ymtdzzz/telemetry-glue/pkg/analyzer/backend"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
)

// AnalysisType represents the type of analysis to be performed
type AnalysisType string

const (
	AnalysisTypeDuration AnalysisType = "duration"
	AnalysisTypeError    AnalysisType = "error"
)

// Analyzer struct that uses an LLMBackend to analyze telemetry data
type Analyzer struct {
	backend  *backend.LLMBackend
	language string
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(config *config.AnalyzerConfig) (*Analyzer, error) {
	backend, err := backend.NewLLMBackend(config)
	if err != nil {
		return nil, err
	}
	return &Analyzer{
		backend: &backend,
	}, nil
}

// AnalyzeDuration generates a report based on the provided telemetry data and prompt
func (a *Analyzer) AnalyzeDuration(ctx context.Context, telemetry *model.Telemetry) (string, error) {
	content, err := generatePrompt(AnalysisTypeDuration, telemetry, a.language)
	if err != nil {
		return "", err
	}
	return (*a.backend).GenerateReport(ctx, content)
}
