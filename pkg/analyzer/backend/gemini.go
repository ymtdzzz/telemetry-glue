package backend

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
)

type Gemini struct {
	llm       *googleai.GoogleAI
	modelName string
}

func NewGemini(ctx context.Context, config *config.GeminiConfig) (*Gemini, error) {
	llm, err := googleai.New(ctx, googleai.WithDefaultModel(config.ModelName), googleai.WithAPIKey(config.APIKey))
	if err != nil {
		return nil, err
	}

	return &Gemini{
		llm:       llm,
		modelName: config.ModelName,
	}, nil
}

func (g *Gemini) GenerateReport(
	ctx context.Context,
	content []llms.MessageContent,
) (string, error) {
	return getGeneratedContent(ctx, g.llm, content)
}
