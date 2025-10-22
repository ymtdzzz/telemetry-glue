package backend

import (
	"context"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
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
		content []llms.MessageContent,
	) (string, error)
}

// NewLLMBackend creates a new LLMBackend based on the provided configuration
func NewLLMBackend(cfg *config.AnalyzerConfig) (LLMBackend, error) {
	if cfg.Ollama != nil {
		return NewOllama(cfg.Ollama)
	}
	return nil, errors.New("no valid LLM backend configuration found")
}
