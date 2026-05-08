// Package throttle provides per-service output throttling based on diff volume.
package throttle

import (
	"errors"
	"sync"
	"time"
)

// Throttle tracks per-service diff counts within a sliding window and
// suppresses output once a configured max is reached.
type Throttle struct {
	mu       sync.Mutex
	maxDiffs int
	window   time.Duration
	buckets  map[string][]time.Time
	now      func() time.Time
}

// New creates a Throttle that allows at most maxDiffs diffs per service
// within the given window duration.
func New(maxDiffs int, window time.Duration) (*Throttle, error) {
	if maxDiffs <= 0 {
		return nil, errors.New("throttle: maxDiffs must be greater than zero")
	}
	if window <= 0 {
		return nil, errors.New("throttle: window must be greater than zero")
	}
	return &Throttle{
		maxDiffs: maxDiffs,
		window:   window,
		buckets:  make(map[string][]time.Time),
		now:      time.Now,
	}, nil
}

// Allow returns true if the diff for the given service should be passed
// through, or false if it should be suppressed due to throttling.
func (t *Throttle) Allow(service string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	cutoff := now.Add(-t.window)

	times := t.buckets[service]
	// evict entries outside the window
	filtered := times[:0]
	for _, ts := range times {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}

	if len(filtered) >= t.maxDiffs {
		t.buckets[service] = filtered
		return false
	}

	t.buckets[service] = append(filtered, now)
	return true
}

// Reset clears all recorded timestamps for a service.
func (t *Throttle) Reset(service string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.buckets, service)
}
