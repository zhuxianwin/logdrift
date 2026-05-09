// Package truncate provides field-level value truncation for log entries,
// capping string values at a configurable maximum byte length.
package truncate

import (
	"fmt"

	"github.com/yourorg/logdrift/internal/parser"
)

// Rule describes a single truncation rule: which field to truncate and the
// maximum number of bytes to retain.
type Rule struct {
	Field  string
	MaxLen int
}

// Truncator applies a set of truncation rules to log entries.
type Truncator struct {
	rules []Rule
}

// New creates a Truncator from the provided rules. Every rule must specify a
// non-empty field name and a positive MaxLen.
func New(rules []Rule) (*Truncator, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("truncate: at least one rule is required")
	}
	for i, r := range rules {
		if r.Field == "" {
			return nil, fmt.Errorf("truncate: rule[%d]: field must not be empty", i)
		}
		if r.MaxLen <= 0 {
			return nil, fmt.Errorf("truncate: rule[%d]: max_len must be positive, got %d", i, r.MaxLen)
		}
	}
	return &Truncator{rules: rules}, nil
}

// Apply returns a shallow copy of entry with string fields truncated according
// to the configured rules. Fields that are absent or not strings are left
// untouched.
func (t *Truncator) Apply(entry parser.Entry) parser.Entry {
	out := entry
	fields := make(map[string]any, len(entry.Fields))
	for k, v := range entry.Fields {
		fields[k] = v
	}
	for _, r := range t.rules {
		v, ok := fields[r.Field]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		if len(s) > r.MaxLen {
			fields[r.Field] = s[:r.MaxLen]
		}
	}
	out.Fields = fields
	return out
}
