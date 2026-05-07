package replay

import (
	"errors"
	"fmt"
)

// ServiceReplay holds the replay configuration for a single service.
type ServiceReplay struct {
	// Name is the logical service name (must match a service in the main config).
	Name string `yaml:"name"`
	// File is the path to the log file to replay.
	File string `yaml:"file"`
}

// Config holds global replay settings.
type Config struct {
	// Speed is the playback multiplier. 0 or negative means no delay.
	Speed    float64         `yaml:"speed"`
	Services []ServiceReplay `yaml:"services"`
}

// BuildOptions converts a Config into Options for use with Run/OpenFile.
func BuildOptions(cfg Config) Options {
	if cfg.Speed <= 0 {
		return Options{NoDelay: true}
	}
	return Options{Speed: cfg.Speed}
}

// Validate checks that the Config is well-formed.
func Validate(cfg Config) error {
	if len(cfg.Services) == 0 {
		return errors.New("replay: at least one service must be specified")
	}

	seen := make(map[string]struct{}, len(cfg.Services))
	for i, svc := range cfg.Services {
		if svc.Name == "" {
			return fmt.Errorf("replay: service[%d]: name must not be empty", i)
		}
		if svc.File == "" {
			return fmt.Errorf("replay: service %q: file must not be empty", svc.Name)
		}
		if _, dup := seen[svc.Name]; dup {
			return fmt.Errorf("replay: duplicate service name %q", svc.Name)
		}
		seen[svc.Name] = struct{}{}
	}

	if cfg.Speed < 0 {
		return fmt.Errorf("replay: speed must be >= 0, got %g", cfg.Speed)
	}

	return nil
}
