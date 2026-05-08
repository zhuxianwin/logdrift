// Package labelmap provides field-value label injection for log entries,
// allowing static or derived labels to be attached to parsed log entries
// before they reach the differ or formatter stages.
package labelmap

import (
	"fmt"
	"regexp"

	"github.com/yourorg/logdrift/internal/parser"
)

// Rule describes a single label injection: if SourceField matches Pattern,
// set DestField to Label.
type Rule struct {
	SourceField string
	Pattern     *regexp.Regexp
	DestField   string
	Label       string
}

// Mapper holds a set of label rules and applies them to log entries.
type Mapper struct {
	rules []Rule
}

// New constructs a Mapper from the provided rules, returning an error if any
// rule is misconfigured.
func New(rules []Rule) (*Mapper, error) {
	for i, r := range rules {
		if r.SourceField == "" {
			return nil, fmt.Errorf("rule %d: source_field must not be empty", i)
		}
		if r.DestField == "" {
			return nil, fmt.Errorf("rule %d: dest_field must not be empty", i)
		}
		if r.Pattern == nil {
			return nil, fmt.Errorf("rule %d: pattern must not be nil", i)
		}
		if r.Label == "" {
			return nil, fmt.Errorf("rule %d: label must not be empty", i)
		}
	}
	return &Mapper{rules: rules}, nil
}

// Apply iterates over all rules and injects matching labels into a copy of the
// entry's Fields map. The original entry is not mutated.
func (m *Mapper) Apply(e parser.Entry) parser.Entry {
	if len(m.rules) == 0 {
		return e
	}

	// shallow-copy fields so we don't mutate the original
	newFields := make(map[string]interface{}, len(e.Fields))
	for k, v := range e.Fields {
		newFields[k] = v
	}

	for _, r := range m.rules {
		val, ok := newFields[r.SourceField]
		if !ok {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		if r.Pattern.MatchString(str) {
			newFields[r.DestField] = r.Label
		}
	}

	e.Fields = newFields
	return e
}
