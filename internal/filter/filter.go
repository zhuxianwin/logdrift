// Package filter provides log entry filtering based on field matchers.
package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/user/logdrift/internal/parser"
)

// Rule defines a single filter condition on a named field.
type Rule struct {
	Field   string
	Pattern string
	re      *regexp.Regexp
}

// Filter holds a compiled set of rules and applies them to log entries.
type Filter struct {
	rules []Rule
}

// New compiles the provided rules and returns a Filter.
// Returns an error if any pattern fails to compile.
func New(rules []Rule) (*Filter, error) {
	compiled := make([]Rule, len(rules))
	for i, r := range rules {
		re, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, err
		}
		compiled[i] = Rule{Field: r.Field, Pattern: r.Pattern, re: re}
	}
	return &Filter{rules: compiled}, nil
}

// Match returns true when the entry satisfies ALL rules.
// A rule matches if the field value (string-coerced) matches the pattern.
// Fields absent from the entry cause the rule to fail.
func (f *Filter) Match(entry parser.Entry) bool {
	for _, rule := range f.rules {
		val, ok := entry.Fields[rule.Field]
		if !ok {
			return false
		}
		s := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", val)))
		if !rule.re.MatchString(s) {
			return false
		}
	}
	return true
}

// MatchAny returns true when the entry satisfies AT LEAST ONE rule.
func (f *Filter) MatchAny(entry parser.Entry) bool {
	for _, rule := range f.rules {
		val, ok := entry.Fields[rule.Field]
		if !ok {
			continue
		}
		s := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", val)))
		if rule.re.MatchString(s) {
			return true
		}
	}
	return false
}
