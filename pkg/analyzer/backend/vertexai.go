package backend

import (
	"context"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
)

type VertexAI struct {
	llm       *vertex.Vertex
	modelName string
}

func NewVertexAI(ctx context.Context, config *config.VertexAIConfig) (*VertexAI, error) {
	llm, err := vertex.New(
		ctx,
		googleai.WithDefaultModel(config.ModelName),
		googleai.WithCloudProject(config.ProjectID),
		googleai.WithCloudLocation(config.Location),
	)
	if err != nil {
		return nil, err
	}

	return &VertexAI{
		llm:       llm,
		modelName: config.ModelName,
	}, nil
}

func (v *VertexAI) GenerateReport(
	ctx context.Context,
	content []llms.MessageContent,
) (string, error) {
	return getGeneratedContent(ctx, v.llm, content)
}
