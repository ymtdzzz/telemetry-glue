package backend

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
)

type Ollama struct {
	llm       *ollama.LLM
	modelName string
}

func NewOllama(config *config.OllamaConfig) (*Ollama, error) {
	llm, err := ollama.New(ollama.WithModel(config.ModelName))
	if err != nil {
		return nil, err
	}

	return &Ollama{
		llm:       llm,
		modelName: config.ModelName,
	}, nil
}

func (o *Ollama) GenerateReport(
	ctx context.Context,
	content []llms.MessageContent,
) (string, error) {
	return getGeneratedContent(ctx, o.llm, content)
}
