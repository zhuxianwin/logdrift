package sampler

import (
	"fmt"
	"time"
)

// ServiceConfig mirrors the sampler section of a service's config entry.
type ServiceConfig struct {
	Enabled bool    `yaml:"enabled"`
	Rate    float64 `yaml:"rate"`
	WindowS int     `yaml:"window_seconds"`
}

// FromServiceConfig constructs a Sampler from a ServiceConfig.
// If cfg.Enabled is false, a pass-through sampler (rate=1.0) is returned.
func FromServiceConfig(cfg ServiceConfig) (*Sampler, error) {
	if !cfg.Enabled {
		return New(Config{Rate: 1.0, Window: time.Second}, nil)
	}
	if cfg.Rate == 0 {
		cfg.Rate = 1.0
	}
	win := time.Duration(cfg.WindowS) * time.Second
	if win <= 0 {
		win = time.Second
	}
	s, err := New(Config{Rate: cfg.Rate, Window: win}, nil)
	if err != nil {
		return nil, fmt.Errorf("sampler config: %w", err)
	}
	return s, nil
}
