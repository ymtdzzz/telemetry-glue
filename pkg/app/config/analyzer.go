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
