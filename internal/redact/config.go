package redact

import (
	"fmt"

	"github.com/yourorg/logdrift/internal/config"
)

// FromConfig builds a Redactor from the service-level redaction config.
// Returns nil, nil when no rules are defined.
func FromConfig(rules []config.RedactRule) (*Redactor, error) {
	if len(rules) == 0 {
		return nil, nil
	}
	rs := make([]Rule, 0, len(rules))
	for i, cr := range rules {
		if cr.Field == "" {
			return nil, fmt.Errorf("redact: config rule %d missing field", i)
		}
		rs = append(rs, Rule{
			Field:   cr.Field,
			Pattern: cr.Pattern,
		})
	}
	return New(rs)
}
