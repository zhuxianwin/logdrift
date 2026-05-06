package ratelimit

import "fmt"

// ServiceConfig mirrors the rate-limit section of a service config entry.
type ServiceConfig struct {
	RatePerSec float64 `yaml:"rate_per_sec"`
	Burst      float64 `yaml:"burst"`
}

// DefaultRatePerSec is used when no rate is specified in config.
const DefaultRatePerSec = 10.0

// DefaultBurst is used when no burst is specified in config.
const DefaultBurst = 5.0

// FromServiceConfig builds a Limiter from a ServiceConfig, applying
// defaults for zero-value fields.
func FromServiceConfig(cfg ServiceConfig) (*Limiter, error) {
	rate := cfg.RatePerSec
	if rate <= 0 {
		rate = DefaultRatePerSec
	}
	burst := cfg.Burst
	if burst <= 0 {
		burst = DefaultBurst
	}
	l, err := New(rate, burst)
	if err != nil {
		return nil, fmt.Errorf("ratelimit.FromServiceConfig: %w", err)
	}
	return l, nil
}
