package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Service defines a single service to tail.
type Service struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// Config holds the top-level logdrift configuration.
type Config struct {
	Services   []Service `yaml:"services"`
	WindowSize int       `yaml:"window_size"`
	OutputMode string    `yaml:"output_mode"` // "color" or "plain"
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	setDefaults(&cfg)
	return &cfg, nil
}

func validate(cfg *Config) error {
	if len(cfg.Services) == 0 {
		return fmt.Errorf("config: at least one service must be defined")
	}
	seen := make(map[string]bool)
	for i, svc := range cfg.Services {
		if svc.Name == "" {
			return fmt.Errorf("config: service[%d] missing name", i)
		}
		if svc.Path == "" {
			return fmt.Errorf("config: service %q missing path", svc.Name)
		}
		if seen[svc.Name] {
			return fmt.Errorf("config: duplicate service name %q", svc.Name)
		}
		seen[svc.Name] = true
	}
	if cfg.OutputMode != "" && cfg.OutputMode != "color" && cfg.OutputMode != "plain" {
		return fmt.Errorf("config: output_mode must be \"color\" or \"plain\", got %q", cfg.OutputMode)
	}
	return nil
}

func setDefaults(cfg *Config) {
	if cfg.WindowSize <= 0 {
		cfg.WindowSize = 100
	}
	if cfg.OutputMode == "" {
		cfg.OutputMode = "color"
	}
}
