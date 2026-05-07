package correlation

import (
	"fmt"
	"time"
)

// Config holds the user-facing configuration for correlation tracking.
type Config struct {
	// Field is the JSON field name used as the correlation identifier
	// (e.g. "trace_id", "request_id").
	Field string `yaml:"field"`

	// TTL is how long a group is retained after its last entry, expressed as a
	// Go duration string (e.g. "30s", "2m").
	TTL string `yaml:"ttl"`
}

// FromConfig constructs a Tracker from a Config. If cfg is nil the function
// returns nil without error so callers can treat correlation as optional.
func FromConfig(cfg *Config) (*Tracker, error) {
	if cfg == nil {
		return nil, nil
	}
	if cfg.Field == "" {
		return nil, fmt.Errorf("correlation config: field is required")
	}
	ttlStr := cfg.TTL
	if ttlStr == "" {
		ttlStr = "5m"
	}
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, fmt.Errorf("correlation config: invalid ttl %q: %w", ttlStr, err)
	}
	return New(cfg.Field, ttl)
}
