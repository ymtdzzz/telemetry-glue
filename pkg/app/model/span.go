package model

import (
	"fmt"
	"strings"
)

// Span represents single span
type Span map[string]any

// Spans represents spans
type Spans []Span

func (ss *Spans) AsCSV() (string, error) {
	keyMap := map[string]bool{}

	for _, s := range *ss {
		for k := range s {
			keyMap[k] = true
		}
	}

	csvData := strings.Builder{}

	headers := []string{}
	for k := range keyMap {
		headers = append(headers, k)
	}

	csvData.WriteString(strings.Join(headers, ",") + "\n")

	for _, span := range *ss {
		row := []string{}
		for _, header := range headers {
			if val, ok := span[header]; ok {
				row = append(row, fmt.Sprintf("%v", val))
			} else {
				row = append(row, "")
			}
		}
		csvData.WriteString(strings.Join(row, ",") + "\n")
	}

	return csvData.String(), nil
}
