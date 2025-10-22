package backend

import (
	"context"
	"log"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"golang.org/x/sync/errgroup"
)

type Ollama struct {
	llm *ollama.LLM
}

func NewOllama(config *config.OllamaConfig) (*Ollama, error) {
	llm, err := ollama.New(ollama.WithModel(config.ModelName))
	if err != nil {
		return nil, err
	}

	return &Ollama{
		llm: llm,
	}, nil
}

func (o *Ollama) GenerateReport(
	ctx context.Context,
	content []llms.MessageContent,
) (string, error) {
	g := new(errgroup.Group)
	chunks := make(chan string)

	g.Go(func() error {
		defer close(chunks)
		_, err := o.llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			log.Print(string(chunk))
			chunks <- string(chunk)
			return nil
		}))
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return "", err
	}

	var result strings.Builder
	for chunk := range chunks {
		result.WriteString(chunk)
	}

	return result.String(), nil
}
