package backend

import (
	"fmt"
)

// BackendType represents the type of backend
type BackendType string

const (
	NewRelicBackend BackendType = "newrelic"
)

// GetSupportedBackends returns a list of supported backend types
func GetSupportedBackends() []BackendType {
	return []BackendType{
		NewRelicBackend,
	}
}

// ValidateBackendType checks if the given backend type is supported
func ValidateBackendType(backendType string) error {
	for _, supported := range GetSupportedBackends() {
		if string(supported) == backendType {
			return nil
		}
	}
	return fmt.Errorf("unsupported backend type: %s", backendType)
}
