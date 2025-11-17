package config

import "errors"

type BackendType string

const (
	BackendTypeNewRelic BackendType = "newrelic"
)

type GlueConfig struct {
	NewRelic    NewRelicConfig `yaml:"newrelic,omitempty" envPrefix:"NEW_RELIC_"`
	SpanBackend BackendType    `yaml:"span" env:"SPAN_BACKEND"`
	LogBackend  BackendType    `yaml:"log" env:"LOG_BACKEND"`
}

func (c *GlueConfig) hasAnyConfig() bool {
	return c.SpanBackend != "" || c.LogBackend != "" || c.NewRelic.HasAnyConfig()
}

func (c *GlueConfig) validate() error {
	if c.SpanBackend == BackendTypeNewRelic || c.LogBackend == BackendTypeNewRelic {
		if !c.NewRelic.HasAnyConfig() {
			return errors.New("the New Relic configuration is required for the selected backend")
		}
		if err := c.NewRelic.validate(); err != nil {
			return err
		}
	}

	if c.SpanBackend == "" && c.LogBackend == "" {
		return errors.New("at least one backend must be configured")
	}

	return nil
}

type NewRelicConfig struct {
	APIKey    string `yaml:"api_key" env:"API_KEY"`
	AccountID int    `yaml:"account_id" env:"ACCOUNT_ID"`
}

func (c *NewRelicConfig) HasAnyConfig() bool {
	return c.APIKey != "" || c.AccountID != 0
}

func (c *NewRelicConfig) validate() error {
	if c.APIKey == "" {
		return errors.New("the New Relic API key is required")
	}
	if c.AccountID == 0 {
		return errors.New("the New Relic Account ID is required")
	}
	return nil
}
