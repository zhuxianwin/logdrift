package alerting

import (
	"fmt"
	"io"
	"time"

	"github.com/user/logdrift/internal/differ"
)

// Threshold defines conditions that trigger an alert.
type Threshold struct {
	// MinDeltaCount is the minimum number of differing fields to trigger.
	MinDeltaCount int
	// Window is how long a breach must persist before alerting.
	Window time.Duration
}

// Alert represents a fired alert event.
type Alert struct {
	FiredAt    time.Time
	DeltaCount int
	Message    string
}

// Alerter watches diff output and fires alerts when thresholds are breached.
type Alerter struct {
	threshold  Threshold
	writer     io.Writer
	breachSince *time.Time
	now        func() time.Time
}

// New creates an Alerter with the given threshold writing alerts to w.
func New(t Threshold, w io.Writer) (*Alerter, error) {
	if t.MinDeltaCount < 1 {
		return nil, fmt.Errorf("alerting: MinDeltaCount must be >= 1, got %d", t.MinDeltaCount)
	}
	if t.Window < 0 {
		return nil, fmt.Errorf("alerting: Window must be non-negative")
	}
	return &Alerter{
		threshold: t,
		writer:    w,
		now:       time.Now,
	}, nil
}

// Evaluate checks a slice of deltas against the threshold.
// It returns a non-nil Alert when one is fired, otherwise nil.
func (a *Alerter) Evaluate(deltas []differ.Delta) *Alert {
	count := len(deltas)
	now := a.now()

	if count >= a.threshold.MinDeltaCount {
		if a.breachSince == nil {
			a.breachSince = &now
		}
		if now.Sub(*a.breachSince) >= a.threshold.Window {
			msg := fmt.Sprintf("[ALERT] %s drift detected: %d field(s) differ (threshold: %d)",
				now.Format(time.RFC3339), count, a.threshold.MinDeltaCount)
			fmt.Fprintln(a.writer, msg)
			return &Alert{FiredAt: now, DeltaCount: count, Message: msg}
		}
	} else {
		a.breachSince = nil
	}
	return nil
}
