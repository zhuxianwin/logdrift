// Package aggregator groups log entries by a key field and computes
// aggregate statistics (count, first/last seen) over a sliding window.
package aggregator

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/logdrift/internal/parser"
)

// Group holds aggregated data for a single key value.
type Group struct {
	Key       string
	Count     int
	FirstSeen time.Time
	LastSeen  time.Time
	Services  map[string]int
}

// Aggregator buckets parsed log entries by a configurable field.
type Aggregator struct {
	mu      sync.Mutex
	field   string
	window  time.Duration
	groups  map[string]*Group
	nowFunc func() time.Time
}

// New creates an Aggregator that groups entries by field over the given window.
// Returns an error if field is empty or window is non-positive.
func New(field string, window time.Duration) (*Aggregator, error) {
	if field == "" {
		return nil, fmt.Errorf("aggregator: field must not be empty")
	}
	if window <= 0 {
		return nil, fmt.Errorf("aggregator: window must be positive, got %s", window)
	}
	return &Aggregator{
		field:   field,
		window:  window,
		groups:  make(map[string]*Group),
		nowFunc: time.Now,
	}, nil
}

// Add records a parsed entry into the appropriate group, evicting stale groups.
func (a *Aggregator) Add(entry parser.Entry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.evict()

	val, ok := entry.Fields[a.field]
	if !ok {
		return
	}
	key := fmt.Sprintf("%v", val)

	g, exists := a.groups[key]
	if !exists {
		g = &Group{
			Key:       key,
			FirstSeen: entry.Timestamp,
			Services:  make(map[string]int),
		}
		a.groups[key] = g
	}
	g.Count++
	g.LastSeen = entry.Timestamp
	g.Services[entry.Service]++
}

// Groups returns a snapshot of all current (non-evicted) groups.
func (a *Aggregator) Groups() []Group {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.evict()

	out := make([]Group, 0, len(a.groups))
	for _, g := range a.groups {
		copy := *g
		copy.Services = make(map[string]int, len(g.Services))
		for k, v := range g.Services {
			copy.Services[k] = v
		}
		out = append(out, copy)
	}
	return out
}

// evict removes groups whose LastSeen is outside the window. Must be called with lock held.
func (a *Aggregator) evict() {
	cutoff := a.nowFunc().Add(-a.window)
	for key, g := range a.groups {
		if g.LastSeen.Before(cutoff) {
			delete(a.groups, key)
		}
	}
}
