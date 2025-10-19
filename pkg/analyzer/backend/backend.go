package backend

import (
	"context"
	"errors"

	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
)

// AnalysisType represents the type of analysis to perform
type AnalysisType string

const (
	AnalysisTypeDuration AnalysisType = "duration"
	AnalysisTypeError    AnalysisType = "error"
)

// LLMBackend defines the interface for telemetry analyzers
type LLMBackend interface {
	GenerateReport(
		ctx context.Context,
		telemetry *model.Telemetry,
		prompt string,
		analysisType AnalysisType,
	) (string, error)
}

// NewLLMBackend creates a new LLMBackend based on the provided configuration
func NewLLMBackend(config *config.AnalyzerConfig) (LLMBackend, error) {
	if config.VertexAI != nil {
		return NewVertexAI(config.VertexAI)
	}
	return nil, errors.New("no valid LLM backend configuration found")
}
