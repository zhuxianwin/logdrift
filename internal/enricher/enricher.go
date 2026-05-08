package enricher

import (
	"fmt"
	"strings"

	"github.com/user/logdrift/internal/parser"
)

// Rule defines a single enrichment: copy or transform a field into a new field.
type Rule struct {
	// SourceField is the field to read from the log entry.
	SourceField string
	// DestField is the field to write the enriched value into.
	DestField string
	// Prefix is an optional static prefix prepended to the value.
	Prefix string
	// Uppercase transforms the value to upper case when true.
	Uppercase bool
}

// Enricher applies a set of rules to log entries, adding derived fields.
type Enricher struct {
	rules []Rule
}

// New creates an Enricher from the provided rules.
// Returns an error if any rule has an empty SourceField or DestField.
func New(rules []Rule) (*Enricher, error) {
	for i, r := range rules {
		if strings.TrimSpace(r.SourceField) == "" {
			return nil, fmt.Errorf("rule %d: source_field must not be empty", i)
		}
		if strings.TrimSpace(r.DestField) == "" {
			return nil, fmt.Errorf("rule %d: dest_field must not be empty", i)
		}
	}
	return &Enricher{rules: rules}, nil
}

// Apply returns a new parser.Entry with enriched fields added.
// Original fields are never overwritten; if DestField already exists it is skipped.
func (e *Enricher) Apply(entry parser.Entry) parser.Entry {
	if len(e.rules) == 0 {
		return entry
	}

	// Shallow-copy the fields map so we don't mutate the original.
	newFields := make(map[string]interface{}, len(entry.Fields)+len(e.rules))
	for k, v := range entry.Fields {
		newFields[k] = v
	}

	for _, r := range e.rules {
		if _, exists := newFields[r.DestField]; exists {
			continue
		}
		raw, ok := newFields[r.SourceField]
		if !ok {
			continue
		}
		val := fmt.Sprintf("%v", raw)
		if r.Uppercase {
			val = strings.ToUpper(val)
		}
		if r.Prefix != "" {
			val = r.Prefix + val
		}
		newFields[r.DestField] = val
	}

	return parser.Entry{
		Timestamp: entry.Timestamp,
		Level:     entry.Level,
		Message:   entry.Message,
		Service:   entry.Service,
		Fields:    newFields,
	}
}
