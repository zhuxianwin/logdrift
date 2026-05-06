package alerting

import (
	"fmt"
	"io"
	"time"

	"github.com/user/logdrift/internal/config"
)

// FromConfig builds an Alerter from the application config.
// Returns nil, nil when alerting is not configured.
func FromConfig(cfg *config.Config, w io.Writer) (*Alerter, error) {
	if cfg.Alert == nil {
		return nil, nil
	}

	if cfg.Alert.MinDeltaCount == 0 {
		cfg.Alert.MinDeltaCount = 1
	}

	win, err := time.ParseDuration(cfg.Alert.Window)
	if err != nil {
		return nil, fmt.Errorf("alerting: invalid window %q: %w", cfg.Alert.Window, err)
	}

	return New(Threshold{
		MinDeltaCount: cfg.Alert.MinDeltaCount,
		Window:        win,
	}, w)
}
