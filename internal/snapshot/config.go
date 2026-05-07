package snapshot

import (
	"fmt"
	"strings"
)

// Options controls snapshot capture and persistence behaviour.
type Options struct {
	// Label is a human-readable name for the snapshot.
	Label string
	// OutputPath is the file path where the snapshot will be saved.
	// If empty, the snapshot is kept in memory only.
	OutputPath string
	// MaxEntriesPerService caps how many entries are stored per service.
	// 0 means unlimited.
	MaxEntriesPerService int
}

// Validate checks that the Options are consistent and returns an error
// describing the first problem found.
func (o Options) Validate() error {
	if strings.TrimSpace(o.Label) == "" {
		return fmt.Errorf("snapshot: label must not be empty")
	}
	if o.MaxEntriesPerService < 0 {
		return fmt.Errorf("snapshot: max_entries_per_service must be >= 0, got %d", o.MaxEntriesPerService)
	}
	return nil
}

// BuildOptions constructs an Options from a raw config map, applying
// defaults for any missing keys.
func BuildOptions(cfg map[string]any) (Options, error) {
	opts := Options{
		Label:                "default",
		MaxEntriesPerService: 0,
	}
	if v, ok := cfg["label"].(string); ok && strings.TrimSpace(v) != "" {
		opts.Label = v
	}
	if v, ok := cfg["output_path"].(string); ok {
		opts.OutputPath = v
	}
	if v, ok := cfg["max_entries_per_service"].(int); ok {
		opts.MaxEntriesPerService = v
	}
	if err := opts.Validate(); err != nil {
		return Options{}, err
	}
	return opts, nil
}
