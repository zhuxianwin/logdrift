package schema

import (
	"fmt"
	"sort"

	"github.com/robinmuhia/logdrift/internal/parser"
)

// FieldStats tracks observed values and occurrence count for a single field.
type FieldStats struct {
	Count  int
	Values map[string]int
}

// Schema accumulates field presence and value distribution across log entries.
type Schema struct {
	fields map[string]*FieldStats
	total  int
}

// New returns an empty Schema ready for observation.
func New() *Schema {
	return &Schema{
		fields: make(map[string]*FieldStats),
	}
}

// Observe records all raw fields from a parsed log entry into the schema.
func (s *Schema) Observe(entry parser.Entry) {
	s.total++
	for k, v := range entry.Raw {
		fs, ok := s.fields[k]
		if !ok {
			fs = &FieldStats{Values: make(map[string]int)}
			s.fields[k] = fs
		}
		fs.Count++
		val := fmt.Sprintf("%v", v)
		fs.Values[val]++
	}
}

// Fields returns a sorted list of all observed field names.
func (s *Schema) Fields() []string {
	keys := make([]string, 0, len(s.fields))
	for k := range s.fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Stats returns FieldStats for a given field name, and whether it was found.
func (s *Schema) Stats(field string) (FieldStats, bool) {
	fs, ok := s.fields[field]
	if !ok {
		return FieldStats{}, false
	}
	return *fs, true
}

// Total returns the number of entries observed.
func (s *Schema) Total() int {
	return s.total
}

// Coverage returns the fraction of entries (0.0–1.0) that contain the given field.
func (s *Schema) Coverage(field string) float64 {
	if s.total == 0 {
		return 0
	}
	fs, ok := s.fields[field]
	if !ok {
		return 0
	}
	return float64(fs.Count) / float64(s.total)
}
