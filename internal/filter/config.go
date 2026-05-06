package filter

import (
	"fmt"
	"strings"
)

// RuleConfig is the serialisable form of a filter rule, used in config files.
type RuleConfig struct {
	Field   string `yaml:"field"`
	Pattern string `yaml:"pattern"`
	Mode    string `yaml:"mode"` // "all" (default) or "any"
}

// FromConfig converts a slice of RuleConfig into a compiled Filter.
// Mode is derived from the first rule that specifies one; defaults to "all".
func FromConfig(cfgs []RuleConfig) (*Filter, string, error) {
	if len(cfgs) == 0 {
		f, _ := New(nil)
		return f, "all", nil
	}

	mode := "all"
	for _, c := range cfgs {
		if m := strings.ToLower(strings.TrimSpace(c.Mode)); m == "any" || m == "all" {
			mode = m
			break
		}
	}

	rules := make([]Rule, len(cfgs))
	for i, c := range cfgs {
		if strings.TrimSpace(c.Field) == "" {
			return nil, "", fmt.Errorf("filter rule %d: field must not be empty", i)
		}
		if strings.TrimSpace(c.Pattern) == "" {
			return nil, "", fmt.Errorf("filter rule %d: pattern must not be empty", i)
		}
		rules[i] = Rule{Field: c.Field, Pattern: c.Pattern}
	}

	f, err := New(rules)
	if err != nil {
		return nil, "", err
	}
	return f, mode, nil
}
