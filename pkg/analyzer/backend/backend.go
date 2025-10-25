package backend

import (
	"context"
	"errors"
	"strings"

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
	if cfg.Gemini != nil {
		return NewGemini(context.Background(), cfg.Gemini)
	}
	return nil, errors.New("no valid LLM backend configuration found")
}

func getGeneratedContent(
	ctx context.Context,
	llm llms.Model,
	content []llms.MessageContent,
) (string, error) {
	chunks := make(chan string)
	errChan := make(chan error, 1)
	var result strings.Builder

	go func() {
		defer close(chunks)
		_, err := llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			select {
			case chunks <- string(chunk):
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		}))
		errChan <- err
	}()

	for chunk := range chunks {
		result.WriteString(chunk)
	}

	if err := <-errChan; err != nil {
		return "", err
	}

	return result.String(), nil
}
