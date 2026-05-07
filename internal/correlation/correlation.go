// Package correlation groups log entries across services by a shared trace or
// request ID, making it easy to compare the full lifecycle of a single request.
package correlation

import (
	"fmt"
	"sync"
	"time"

	"github.com/user/logdrift/internal/parser"
)

// Group holds all log entries that share the same correlation ID.
type Group struct {
	ID      string
	Entries map[string][]parser.Entry // keyed by service name
	First   time.Time
	Last    time.Time
}

// Tracker accumulates entries by correlation ID within a sliding TTL window.
type Tracker struct {
	mu    sync.Mutex
	field string
	ttl   time.Duration
	groups map[string]*Group
}

// New creates a Tracker that reads the given field as the correlation ID and
// evicts groups older than ttl.
func New(field string, ttl time.Duration) (*Tracker, error) {
	if field == "" {
		return nil, fmt.Errorf("correlation: field name must not be empty")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("correlation: ttl must be positive")
	}
	return &Tracker{
		field:  field,
		ttl:    ttl,
		groups: make(map[string]*Group),
	}, nil
}

// Add records an entry under its correlation ID. It returns the ID that was
// used, or an empty string if the entry has no correlation field.
func (t *Tracker) Add(service string, e parser.Entry) string {
	id, ok := e.Fields[t.field].(string)
	if !ok || id == "" {
		return ""
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	g, exists := t.groups[id]
	if !exists {
		g = &Group{
			ID:      id,
			Entries: make(map[string][]parser.Entry),
			First:   e.Timestamp,
		}
		t.groups[id] = g
	}
	g.Entries[service] = append(g.Entries[service], e)
	if e.Timestamp.After(g.Last) {
		g.Last = e.Timestamp
	}
	return id
}

// Get returns the group for the given ID, or nil if it does not exist.
func (t *Tracker) Get(id string) *Group {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.groups[id]
}

// Evict removes groups whose last-seen timestamp is older than ttl from now.
func (t *Tracker) Evict(now time.Time) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	cutoff := now.Add(-t.ttl)
	removed := 0
	for id, g := range t.groups {
		if g.Last.Before(cutoff) {
			delete(t.groups, id)
			removed++
		}
	}
	return removed
}

// Len returns the number of active groups.
func (t *Tracker) Len() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.groups)
}
