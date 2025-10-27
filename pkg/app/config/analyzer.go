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
	Language string          `yaml:"language" env:"LANGUAGE"` // en, ja
	Ollama   *OllamaConfig   `yaml:"ollama,omitempty" envPrefix:"OLLAMA_"`
	Gemini   *GeminiConfig   `yaml:"gemini,omitempty" envPrefix:"GEMINI_"`
	VertexAI *VertexAIConfig `yaml:"vertex_ai,omitempty" envPrefix:"VERTEX_AI_"`
}

func (c *AnalyzerConfig) validate() error {
	if c.Language == "" {
		return errors.New("analyzer language is required")
	}
	if c.Language != string(LanguageEnglish) && c.Language != string(LanguageJapanese) {
		return errors.New("unsupported language")
	}

	if c.Ollama != nil {
		return c.Ollama.validate()
	}

	if c.Gemini != nil {
		return c.Gemini.validate()
	}

	if c.VertexAI != nil {
		return c.VertexAI.validate()
	}

	return errors.New("no valid analyzer backend configuration found")
}

type OllamaConfig struct {
	ModelName string `yaml:"model_name" env:"MODEL_NAME"`
}

func (c *OllamaConfig) validate() error {
	if c.ModelName == "" {
		return errors.New("Ollama Model Name is required")
	}
	return nil
}

type GeminiConfig struct {
	ModelName string `yaml:"model_name" env:"mODEL_NAME"`
	APIKey    string `yaml:"api_key" env:"API_KEY"`
}

func (c *GeminiConfig) validate() error {
	if c.ModelName == "" {
		return errors.New("Gemini Model Name is required")
	}
	if c.APIKey == "" {
		return errors.New("Gemini API Key is required")
	}
	return nil
}

type VertexAIConfig struct {
	ModelName string `yaml:"model_name" env:"MODEL_NAME"`
	ProjectID string `yaml:"project_id" env:"PROJECT_ID"`
	Location  string `yaml:"location" env:"LOCATION"`
}

func (c *VertexAIConfig) validate() error {
	if c.ModelName == "" {
		return errors.New("Vertex AI Model Name is required")
	}
	if c.ProjectID == "" {
		return errors.New("Vertex AI Project ID is required")
	}
	if c.Location == "" {
		return errors.New("Vertex AI Location is required")
	}
	return nil
}
