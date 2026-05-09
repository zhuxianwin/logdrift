package ratelimit

import (
	"fmt"
	"time"
)

// ServiceConfig mirrors the subset of config.ServiceConfig relevant to rate limiting.
type ServiceConfig interface {
	GetRateLimit() float64
	GetBurst() int
}

// FromServiceConfig constructs a Limiter from a service configuration.
// Returns nil if the service has no rate-limit configured (rate == 0).
func FromServiceConfig(cfg ServiceConfig) (*Limiter, error) {
	rate := cfg.GetRateLimit()
	if rate == 0 {
		return nil, nil
	}

	burst := cfg.GetBurst()
	if burst <= 0 {
		burst = 1
	}

	rl, err := New(rate, burst, func() time.Time { return time.Now() })
	if err != nil {
		return nil, fmt.Errorf("ratelimit: build from config: %w", err)
	}
	return rl, nil
}
