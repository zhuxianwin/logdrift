package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

// Deduplicator suppresses repeated identical log entries within a rolling window.
type Deduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
	nowFunc func() time.Time
}

// New creates a Deduplicator that suppresses duplicate entries seen within window.
func New(window time.Duration) (*Deduplicator, error) {
	if window <= 0 {
		return nil, fmt.Errorf("dedup: window must be positive, got %v", window)
	}
	return &Deduplicator{
		seen:    make(map[string]time.Time),
		window:  window,
		nowFunc: time.Now,
	}, nil
}

// IsDuplicate returns true if an identical entry was seen within the window.
// It also evicts stale entries on each call to bound memory usage.
func (d *Deduplicator) IsDuplicate(entry parser.Entry) bool {
	key := fingerprint(entry)
	now := d.nowFunc()

	d.mu.Lock()
	defer d.mu.Unlock()

	d.evict(now)

	if _, ok := d.seen[key]; ok {
		return true
	}
	d.seen[key] = now
	return false
}

// evict removes entries older than the dedup window. Must be called with mu held.
func (d *Deduplicator) evict(now time.Time) {
	cutoff := now.Add(-d.window)
	for k, t := range d.seen {
		if t.Before(cutoff) {
			delete(d.seen, k)
		}
	}
}

// fingerprint produces a stable hash of an entry's service + raw fields.
func fingerprint(entry parser.Entry) string {
	type fp struct {
		Service string
		Fields  map[string]any
	}
	v := fp{Service: entry.Service, Fields: entry.Fields}
	b, _ := json.Marshal(v)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
