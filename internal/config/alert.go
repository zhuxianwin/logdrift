package config

// AlertConfig holds threshold settings for drift alerting.
type AlertConfig struct {
	// MinDeltaCount is the minimum number of differing fields
	// that must be present to consider a breach active.
	MinDeltaCount int `yaml:"min_delta_count"`

	// Window is a Go duration string (e.g. "5s", "1m") that
	// specifies how long a breach must persist before an alert fires.
	Window string `yaml:"window"`
}

// validateAlert checks AlertConfig fields when present.
func validateAlert(a *AlertConfig) error {
	if a == nil {
		return nil
	}
	if a.Window == "" {
		a.Window = "0s"
	}
	return nil
}
