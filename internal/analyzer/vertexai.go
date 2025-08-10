package analyzer

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// LLMProvider interface for different LLM providers
type LLMProvider interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
	Close() error
}

// VertexAIProvider implements LLMProvider for Google Vertex AI using the official SDK
type VertexAIProvider struct {
	client    *genai.Client
	projectID string
	location  string
	model     string
}

// NewVertexAIProvider creates a new Vertex AI provider using the official SDK
func NewVertexAIProvider(projectID, location, model string) (*VertexAIProvider, error) {
	ctx := context.Background()

	// Create Vertex AI client with new SDK configuration
	config := &genai.ClientConfig{
		Project:  projectID,
		Location: location,
		Backend:  genai.BackendVertexAI,
	}

	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
	}

	return &VertexAIProvider{
		client:    client,
		projectID: projectID,
		location:  location,
		model:     model,
	}, nil
}

// GenerateContent generates content using Vertex AI Gemini SDK
func (p *VertexAIProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
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

	// Configure generation parameters
	config := &genai.GenerateContentConfig{
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

	// Generate content using the new SDK
	resp, err := p.client.Models.GenerateContent(ctx, p.model, contents, config)
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

// Close closes the Vertex AI client
func (p *VertexAIProvider) Close() error {
	// The new SDK doesn't require explicit client closing
	return nil
}

// MockProvider implements LLMProvider for testing purposes
type MockProvider struct {
	response string
}

// NewMockProvider creates a new mock provider for testing
func NewMockProvider(response string) *MockProvider {
	return &MockProvider{response: response}
}

// GenerateContent returns the mock response
func (p *MockProvider) GenerateContent(ctx context.Context, prompt string) (string, error) {
	return p.response, nil
}

// Close closes the mock provider (no-op)
func (p *MockProvider) Close() error {
	return nil
}

// AnalysisResult represents the result of an LLM analysis
type AnalysisResult struct {
	AnalysisType string `json:"analysis_type"`
	Summary      string `json:"summary"`
	Content      string `json:"content"`
	Provider     string `json:"provider"`
	Model        string `json:"model"`
}

// Analyzer performs LLM-based analysis on telemetry data
type Analyzer struct {
	provider     LLMProvider
	promptGen    *PromptGenerator
	providerName string
	modelName    string
}

// NewAnalyzer creates a new analyzer with the specified provider
func NewAnalyzer(provider LLMProvider, providerName, modelName string) *Analyzer {
	return &Analyzer{
		provider:     provider,
		promptGen:    NewPromptGenerator(),
		providerName: providerName,
		modelName:    modelName,
	}
}

// Analyze performs analysis on the combined telemetry data
func (a *Analyzer) Analyze(ctx context.Context, analysisType AnalysisType, data *CombinedData) (*AnalysisResult, error) {
	return a.AnalyzeWithLanguage(ctx, analysisType, data, "en")
}

// AnalyzeWithLanguage performs analysis on the combined telemetry data with language specification
func (a *Analyzer) AnalyzeWithLanguage(ctx context.Context, analysisType AnalysisType, data *CombinedData, language string) (*AnalysisResult, error) {
	// Generate the prompt
	prompt, err := a.promptGen.GeneratePrompt(analysisType, data, language)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prompt: %w", err)
	}

	// Generate content using the LLM
	content, err := a.provider.GenerateContent(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate analysis: %w", err)
	}

	// Create the result
	result := &AnalysisResult{
		AnalysisType: string(analysisType),
		Summary:      data.Summary(),
		Content:      content,
		Provider:     a.providerName,
		Model:        a.modelName,
	}

	return result, nil
}

// Close closes the analyzer and its provider
func (a *Analyzer) Close() error {
	return a.provider.Close()
}
