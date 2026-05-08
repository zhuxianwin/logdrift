package routing

import "fmt"

// RuleConfig holds the raw configuration for a single routing rule.
type RuleConfig struct {
	Field   string `yaml:"field"`
	Pattern string `yaml:"pattern"`
	Channel string `yaml:"channel"`
}

// FromConfig constructs a Router from a slice of RuleConfig entries.
// Returns nil and no error when cfg is empty, signalling routing is disabled.
func FromConfig(cfg []RuleConfig) (*Router, error) {
	if len(cfg) == 0 {
		return nil, nil
	}
	rules := make([]Rule, 0, len(cfg))
	for i, c := range cfg {
		if c.Field == "" {
			return nil, fmt.Errorf("routing rule %d: field is required", i)
		}
		if c.Pattern == "" {
			return nil, fmt.Errorf("routing rule %d: pattern is required", i)
		}
		if c.Channel == "" {
			return nil, fmt.Errorf("routing rule %d: channel is required", i)
		}
		rules = append(rules, Rule{
			Field:   c.Field,
			Pattern: c.Pattern,
			Channel: c.Channel,
		})
	}
	return New(rules)
}
