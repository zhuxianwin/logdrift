// Package redact provides field-level redaction of sensitive values in parsed log entries.
package redact

import (
	"regexp"
	"strings"

	"github.com/yourorg/logdrift/internal/parser"
)

// Rule defines a single redaction rule: a field name and an optional pattern.
// If Pattern is empty, the entire field value is redacted.
// If Pattern is set, only matching substrings are replaced.
type Rule struct {
	Field   string
	Pattern string
	re      *regexp.Regexp
}

// Redactor applies a set of redaction rules to log entries.
type Redactor struct {
	rules []Rule
}

const redacted = "[REDACTED]"

// New compiles and returns a Redactor for the given rules.
// Returns an error if any pattern fails to compile.
func New(rules []Rule) (*Redactor, error) {
	compiled := make([]Rule, len(rules))
	for i, r := range rules {
		if r.Field == "" {
			return nil, fmt.Errorf("redact: rule %d has empty field name", i)
		}
		compiled[i] = r
		if r.Pattern != "" {
			re, err := regexp.Compile(r.Pattern)
			if err != nil {
				return nil, fmt.Errorf("redact: rule %d invalid pattern %q: %w", i, r.Pattern, err)
			}
			compiled[i].re = re
		}
	}
	return &Redactor{rules: compiled}, nil
}

// Apply returns a copy of the entry with sensitive fields redacted.
// The original entry is not modified.
func (r *Redactor) Apply(e parser.Entry) parser.Entry {
	if len(r.rules) == 0 {
		return e
	}
	out := e
	out.Fields = make(map[string]any, len(e.Fields))
	for k, v := range e.Fields {
		out.Fields[k] = v
	}
	for _, rule := range r.rules {
		val, ok := out.Fields[rule.Field]
		if !ok {
			continue
		}
		s, ok := val.(string)
		if !ok {
			out.Fields[rule.Field] = redacted
			continue
		}
		if rule.re != nil {
			out.Fields[rule.Field] = rule.re.ReplaceAllString(s, redacted)
		} else {
			out.Fields[rule.Field] = strings.Repeat("*", len(s))
		}
	}
	return out
}
