package aggregator

import (
	"fmt"
	"time"
)

// Config holds the configuration for an Aggregator as loaded from the YAML config.
type Config struct {
	Field  string `yaml:"field"`
	Window string `yaml:"window"`
}

// FromConfig constructs an Aggregator from a Config.
// Returns nil, nil when cfg is nil (feature disabled).
func FromConfig(cfg *Config) (*Aggregator, error) {
	if cfg == nil {
		return nil, nil
	}
	if cfg.Field == "" {
		return nil, fmt.Errorf("aggregator config: field is required")
	}
	windowStr := cfg.Window
	if windowStr == "" {
		windowStr = "60s"
	}
	d, err := time.ParseDuration(windowStr)
	if err != nil {
		return nil, fmt.Errorf("aggregator config: invalid window %q: %w", windowStr, err)
	}
	return New(cfg.Field, d)
}
