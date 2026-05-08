package enricher

import "fmt"

// RuleConfig is the serialisable form of a Rule, used in YAML/JSON config.
type RuleConfig struct {
	SourceField string `yaml:"source_field" json:"source_field"`
	DestField   string `yaml:"dest_field"   json:"dest_field"`
	Prefix      string `yaml:"prefix"       json:"prefix"`
	Uppercase   bool   `yaml:"uppercase"    json:"uppercase"`
}

// ServiceEnricherConfig holds the enrichment configuration for one service.
type ServiceEnricherConfig struct {
	Rules []RuleConfig `yaml:"rules" json:"rules"`
}

// FromConfig builds an Enricher from a ServiceEnricherConfig.
// Returns nil, nil when cfg is nil or has no rules.
func FromConfig(cfg *ServiceEnricherConfig) (*Enricher, error) {
	if cfg == nil || len(cfg.Rules) == 0 {
		return nil, nil
	}

	rules := make([]Rule, 0, len(cfg.Rules))
	for i, rc := range cfg.Rules {
		if rc.SourceField == "" {
			return nil, fmt.Errorf("enricher rule %d: source_field is required", i)
		}
		if rc.DestField == "" {
			return nil, fmt.Errorf("enricher rule %d: dest_field is required", i)
		}
		rules = append(rules, Rule{
			SourceField: rc.SourceField,
			DestField:   rc.DestField,
			Prefix:      rc.Prefix,
			Uppercase:   rc.Uppercase,
		})
	}
	return New(rules)
}
