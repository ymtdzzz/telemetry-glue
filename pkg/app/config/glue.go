package config

import "errors"

type BackendType string

const (
	BackendTypeNewRelic BackendType = "newrelic"
)

type GlueConfig struct {
	NewRelic    *NewRelicConfig `yaml:"newrelic,omitempty"`
	SpanBackend BackendType     `yaml:"span"`
	LogBackend  BackendType     `yaml:"log"`
}

func (c *GlueConfig) validate() error {
	if c.SpanBackend == BackendTypeNewRelic || c.LogBackend == BackendTypeNewRelic {
		if c.NewRelic == nil {
			return errors.New("New Relic configuration is required for the selected backend")
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
	APIKey    string `yaml:"api_key" env:"NEW_RELIC_API_KEY"`
	AccountID int    `yaml:"account_id" env:"NEW_RELIC_ACCOUNT_ID"`
}

func (c *NewRelicConfig) validate() error {
	if c.APIKey == "" {
		return errors.New("New Relic API key is required")
	}
	if c.AccountID == 0 {
		return errors.New("New Relic Account ID is required")
	}
	return nil
}
