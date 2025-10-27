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
	Language string        `yaml:"language" env:"LANGUAGE"` // en, ja
	Ollama   *OllamaConfig `yaml:"ollama,omitempty" envPrefix:"OLLAMA_"`
	Gemini   *GeminiConfig `yaml:"gemini,omitempty" envPrefix:"GEMINI_"`
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
	ProjectID string `yaml:"project_id" env:"PROJECT_ID"`
	Location  string `yaml:"location" env:"LOCATION"`
}

func (c *GeminiConfig) validate() error {
	if c.ModelName == "" {
		return errors.New("Gemini Model Name is required")
	}
	if c.APIKey == "" {
		if c.ProjectID == "" || c.Location == "" {
			return errors.New("either API Key or (Project ID and Location) are required for Gemini")
		}
	}
	return nil
}
