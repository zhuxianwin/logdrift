package throttle

import (
	"fmt"
	"time"
)

// ServiceConfig mirrors the throttle section of a service's config entry.
type ServiceConfig struct {
	MaxDiffs int    `yaml:"max_diffs"`
	Window   string `yaml:"window"`
}

// FromServiceConfig builds a Throttle from a ServiceConfig.
// Returns nil, nil when cfg is nil or MaxDiffs is zero (disabled).
func FromServiceConfig(cfg *ServiceConfig) (*Throttle, error) {
	if cfg == nil || cfg.MaxDiffs == 0 {
		return nil, nil
	}

	windowStr := cfg.Window
	if windowStr == "" {
		windowStr = "1m"
	}

	d, err := time.ParseDuration(windowStr)
	if err != nil {
		return nil, fmt.Errorf("throttle: invalid window %q: %w", windowStr, err)
	}

	return New(cfg.MaxDiffs, d)
}
