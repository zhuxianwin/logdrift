package differ

import (
	"sync"
	"time"

	"github.com/logdrift/internal/parser"
)

// WindowEntry holds a parsed log entry alongside its service name and arrival time.
type WindowEntry struct {
	Service   string
	Entry     parser.Entry
	ReceivedAt time.Time
}

// Window buffers recent log entries per service and pairs them for diffing.
type Window struct {
	mu      sync.Mutex
	entries map[string][]WindowEntry
	ttl     time.Duration
}

// NewWindow creates a Window that discards entries older than ttl.
func NewWindow(ttl time.Duration) *Window {
	return &Window{
		entries: make(map[string][]WindowEntry),
		ttl:     ttl,
	}
}

// Add inserts a new entry for the given service, evicting stale entries first.
func (w *Window) Add(service string, entry parser.Entry) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict()
	w.entries[service] = append(w.entries[service], WindowEntry{
		Service:    service,
		Entry:      entry,
		ReceivedAt: time.Now(),
	})
}

// Latest returns the most recent entry for a service, if any.
func (w *Window) Latest(service string) (WindowEntry, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	entries := w.entries[service]
	if len(entries) == 0 {
		return WindowEntry{}, false
	}
	return entries[len(entries)-1], true
}

// Services returns the list of services currently tracked.
func (w *Window) Services() []string {
	w.mu.Lock()
	defer w.mu.Unlock()
	svcs := make([]string, 0, len(w.entries))
	for k := range w.entries {
		svcs = append(svcs, k)
	}
	return svcs
}

// evict removes entries older than the window TTL. Must be called with mu held.
func (w *Window) evict() {
	cutoff := time.Now().Add(-w.ttl)
	for svc, list := range w.entries {
		filtered := list[:0]
		for _, e := range list {
			if e.ReceivedAt.After(cutoff) {
				filtered = append(filtered, e)
			}
		}
		w.entries[svc] = filtered
	}
}
