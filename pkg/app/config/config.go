package config

import (
	"errors"
	"os"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Glue     *GlueConfig     `yaml:"glue,omitempty"`
	Analyzer *AnalyzerConfig `yaml:"analyzer,omitempty"`
}

func LoadConfig(path string) (*AppConfig, error) {
	cfg := &AppConfig{}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *AppConfig) validate() error {
	if c.Glue == nil {
		return errors.New("glue configuration is required")
	}
	if c.Analyzer == nil {
		return errors.New("analyzer configuration is required")
	}
	if err := c.Glue.validate(); err != nil {
		return err
	}
	if err := c.Analyzer.validate(); err != nil {
		return err
	}
	return nil
}
