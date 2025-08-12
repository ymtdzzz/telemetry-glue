package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Format represents supported output formats
type Format string

const (
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
	FormatTable Format = "table"
)

// Formatter interface for types that can format their output
type Formatter interface {
	Print(format Format) error
}

func printJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}

// ParseFormat parses a format string into a Format enum
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "csv":
		return FormatCSV, nil
	case "table", "":
		return FormatTable, nil
	default:
		return "", fmt.Errorf("unsupported format: %s (supported: json, csv, table)", s)
	}
}
