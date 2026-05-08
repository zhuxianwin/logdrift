// Package baseline records a reference snapshot of log entries per service
// and computes deviations from that baseline on subsequent observations.
package baseline

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

// Deviation describes how a field value differs from the recorded baseline.
type Deviation struct {
	Field    string
	Expected string
	Actual   string
}

// Baseline holds reference field values per service, captured at record time.
type Baseline struct {
	mu        sync.RWMutex
	reference map[string]map[string]string // service -> field -> value
	recordedAt map[string]time.Time
	ttl        time.Duration
}

// New creates a Baseline with the given TTL. After TTL the reference for a
// service is considered stale and Record will refresh it automatically.
func New(ttl time.Duration) (*Baseline, error) {
	if ttl <= 0 {
		return nil, errors.New("baseline: ttl must be positive")
	}
	return &Baseline{
		reference:  make(map[string]map[string]string),
		recordedAt: make(map[string]time.Time),
		ttl:        ttl,
	}, nil
}

// Record stores the fields of entry as the baseline for its service.
// If a baseline already exists and has not expired it is left unchanged.
func (b *Baseline) Record(entry parser.Entry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	svc := entry.Service
	if t, ok := b.recordedAt[svc]; ok && time.Since(t) < b.ttl {
		return
	}

	snap := make(map[string]string, len(entry.Fields))
	for k, v := range entry.Fields {
		snap[k] = fmt.Sprintf("%v", v)
	}
	b.reference[svc] = snap
	b.recordedAt[svc] = time.Now()
}

// Compare returns the deviations between entry's fields and the recorded
// baseline for its service. If no baseline has been recorded an empty slice
// is returned.
func (b *Baseline) Compare(entry parser.Entry) []Deviation {
	b.mu.RLock()
	defer b.mu.RUnlock()

	ref, ok := b.reference[entry.Service]
	if !ok {
		return nil
	}

	var devs []Deviation
	for k, refVal := range ref {
		actual, exists := entry.Fields[k]
		if !exists {
			devs = append(devs, Deviation{Field: k, Expected: refVal, Actual: "<missing>"})
			continue
		}
		actualStr := fmt.Sprintf("%v", actual)
		if actualStr != refVal {
			devs = append(devs, Deviation{Field: k, Expected: refVal, Actual: actualStr})
		}
	}
	return devs
}

// Reset removes the baseline for the given service, forcing the next Record
// call to capture a fresh reference.
func (b *Baseline) Reset(service string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.reference, service)
	delete(b.recordedAt, service)
}
