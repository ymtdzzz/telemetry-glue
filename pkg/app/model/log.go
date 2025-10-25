package model

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jeremywohl/flatten/v2"
)

// Log represents single log entry
type Log struct {
	Timestamp  time.Time      `json:"timestamp"`
	TraceID    string         `json:"trace_id,omitempty"`
	SpanID     string         `json:"span_id,omitempty"`
	Message    string         `json:"message"`
	Attributes map[string]any `json:"attributes,omitempty"`
}

func (l *Log) asMap() (map[string]any, error) {
	dataJSON, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}

	flatStr, err := flatten.FlattenString(string(dataJSON), "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(flatStr), &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Logs represents logs
type Logs []Log

func (ls *Logs) AsCSV() (string, error) {
	attributeKeyMap := map[string]bool{}

	logFlatMaps := []map[string]any{}
	for _, l := range *ls {
		flatMap, err := l.asMap()
		if err != nil {
			return "", err
		}
		logFlatMaps = append(logFlatMaps, flatMap)
		for k := range flatMap {
			if strings.HasPrefix(k, "attributes.") {
				attributeKeyMap[k] = true
			}
		}
	}

	csvData := strings.Builder{}

	headers := []string{
		"timestamp",
		"trace_id",
		"span_id",
		"message",
	}
	for k := range attributeKeyMap {
		headers = append(headers, k)
	}

	csvData.WriteString(strings.Join(headers, ",") + "\n")

	for _, flatMap := range logFlatMaps {
		row := []string{}
		for _, header := range headers {
			if val, ok := flatMap[header]; ok {
				row = append(row, fmt.Sprintf("%v", val))
			} else {
				row = append(row, "")
			}
		}
		csvData.WriteString(strings.Join(row, ",") + "\n")
	}

	return csvData.String(), nil
}
