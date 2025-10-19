package backend

import (
	"context"
	"fmt"

	"github.com/ymtdzzz/telemetry-glue/pkg/app/config"
	"github.com/ymtdzzz/telemetry-glue/pkg/app/model"
	"google.golang.org/genai"
)

type VertexAI struct {
	client    *genai.Client
	projectID string
	location  string
	modelName string
}

func NewVertexAI(config *config.VertexAIConfig) (*VertexAI, error) {
	ctx := context.Background()

	// Create Vertex AI client with new SDK configuration
	cfg := &genai.ClientConfig{
		Project:  config.ProjectID,
		Location: config.Location,
		Backend:  genai.BackendVertexAI,
	}

	client, err := genai.NewClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &VertexAI{
		client:    client,
		projectID: config.ProjectID,
		location:  config.Location,
		modelName: config.ModelName,
	}, nil
}

func (v *VertexAI) GenerateReport(
	ctx context.Context,
	telemetry *model.Telemetry,
	prompt string,
	analysisType AnalysisType,
) (string, error) {
	// Create content parts
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{
					Text: prompt,
				},
			},
		},
	}

	config := v.generationConfig()

	// Generate content using the new SDK
	resp, err := v.client.Models.GenerateContent(ctx, v.modelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	// Check for blocked content
	if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != genai.BlockedReasonUnspecified {
		return "", fmt.Errorf("prompt was blocked: %v", resp.PromptFeedback.BlockReason)
	}

	// Extract the text response
	if len(resp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := resp.Candidates[0]

	// Check if candidate was blocked
	if candidate.FinishReason == genai.FinishReasonSafety {
		return "", fmt.Errorf("response was blocked due to safety concerns")
	}

	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	// Convert parts to text
	var result string
	for _, part := range candidate.Content.Parts {
		result += part.Text
	}

	if result == "" {
		return "", fmt.Errorf("no text content found in response")
	}

	return result, nil
}

func (v *VertexAI) generationConfig() *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		Temperature:     genai.Ptr(float32(0.2)),
		MaxOutputTokens: 2048,
		SafetySettings: []*genai.SafetySetting{
			{
				Category:  genai.HarmCategoryHarassment,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategoryHateSpeech,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategorySexuallyExplicit,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
			{
				Category:  genai.HarmCategoryDangerousContent,
				Threshold: genai.HarmBlockThresholdBlockNone,
			},
		},
	}
}
