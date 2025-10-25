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
	Language string        `yaml:"language"` // en, ja
	Ollama   *OllamaConfig `yaml:"ollama,omitempty"`
	Gemini   *GeminiConfig `yaml:"gemini,omitempty"`
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
	ModelName string `yaml:"model_name"`
}

func (c *OllamaConfig) validate() error {
	if c.ModelName == "" {
		return errors.New("Ollama Model Name is required")
	}
	return nil
}

type GeminiConfig struct {
	ModelName string `yaml:"model_name"`
	APIKey    string `yaml:"api_key" env:"GEMINI_API_KEY"`
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
