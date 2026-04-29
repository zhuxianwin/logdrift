package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

// Entry represents a parsed structured JSON log entry.
type Entry struct {
	Service   string
	Raw       string
	Fields    map[string]interface{}
	Timestamp time.Time
	Level     string
	Message   string
}

// Parse attempts to parse a raw JSON log line into an Entry.
// If the line is not valid JSON, it returns an Entry with only the Raw field set.
func Parse(service, line string) (*Entry, error) {
	entry := &Entry{
		Service: service,
		Raw:     line,
		Fields:  make(map[string]interface{}),
	}

	if err := json.Unmarshal([]byte(line), &entry.Fields); err != nil {
		return entry, fmt.Errorf("non-json line: %w", err)
	}

	entry.Timestamp = extractTime(entry.Fields)
	entry.Level = extractString(entry.Fields, "level", "severity", "lvl")
	entry.Message = extractString(entry.Fields, "message", "msg", "text")

	return entry, nil
}

// extractTime tries common timestamp field names and parses them.
func extractTime(fields map[string]interface{}) time.Time {
	for _, key := range []string{"time", "timestamp", "ts", "@timestamp"} {
		val, ok := fields[key]
		if !ok {
			continue
		}
		switch v := val.(type) {
		case string:
			for _, layout := range []string{time.RFC3339Nano, time.RFC3339, time.RFC1123Z} {
				if t, err := time.Parse(layout, v); err == nil {
					return t
				}
			}
		case float64:
			return time.Unix(int64(v), 0).UTC()
		}
	}
	return time.Time{}
}

// extractString returns the first non-empty string value found for the given keys.
func extractString(fields map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val, ok := fields[key]; ok {
			if s, ok := val.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
