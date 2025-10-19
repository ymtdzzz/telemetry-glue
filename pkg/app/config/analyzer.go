package config

import (
	"errors"
)

type Language string

const (
	LanguageEnglish  Language = "en"
	LanguageJapanese Language = "ja"
)

type AnalyzerConfig struct {
	Language string          `yaml:"language"` // en, ja
	VertexAI *VertexAIConfig `yaml:"vertex_ai,omitempty"`
	Gemini   *GeminiConfig   `yaml:"gemini,omitempty"`
}

func (c *AnalyzerConfig) validate() error {
	if c.Language == "" {
		return errors.New("analyzer language is required")
	}
	if c.Language != string(LanguageEnglish) && c.Language != string(LanguageJapanese) {
		return errors.New("unsupported language")
	}

	if c.VertexAI != nil {
		if err := c.VertexAI.validate(); err != nil {
			return err
		}
	}
	if c.Gemini != nil {
		if err := c.Gemini.validate(); err != nil {
			return err
		}
	}
	return nil
}

type VertexAIConfig struct {
	ProjectID string `yaml:"project_id"`
	Location  string `yaml:"location"`
	ModelName string `yaml:"model_name"`
}

func (c *VertexAIConfig) validate() error {
	if c.ProjectID == "" {
		return errors.New("Vertex AI Project ID is required")
	}
	if c.Location == "" {
		return errors.New("Vertex AI Location is required")
	}
	if c.ModelName == "" {
		return errors.New("Vertex AI Model Name is required")
	}
	return nil
}

type GeminiConfig struct {
	*VertexAIConfig
}
