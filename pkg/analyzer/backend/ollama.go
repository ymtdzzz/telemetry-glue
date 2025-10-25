package backend

import (
	"context"
	"log"
	"strings"

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
	chunks := make(chan string)
	errChan := make(chan error, 1)
	var result strings.Builder

	go func() {
		defer close(chunks)
		_, err := o.llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Print(string(chunk))
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
