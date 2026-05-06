// Package sampler provides rate-based sampling for log entries to reduce
// noise when services emit high-frequency identical log lines.
package sampler

import (
	"fmt"
	"sync"
	"time"
)

// Config holds sampling configuration.
type Config struct {
	// Rate is the fraction of entries to keep (0.0–1.0). 1.0 means keep all.
	Rate float64
	// Window is the duration over which the sample counter resets.
	Window time.Duration
}

// bucket tracks entry counts within a rolling window for a single key.
type bucket struct {
	count   int
	resetAt time.Time
}

// Sampler decides whether a log entry should be forwarded downstream.
type Sampler struct {
	cfg     Config
	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time
}

// New returns a Sampler configured with cfg.
// now may be nil, in which case time.Now is used.
func New(cfg Config, now func() time.Time) (*Sampler, error) {
	if cfg.Rate < 0 || cfg.Rate > 1 {
		return nil, fmt.Errorf("sampler: rate must be between 0.0 and 1.0, got %f", cfg.Rate)
	}
	if cfg.Window <= 0 {
		cfg.Window = time.Second
	}
	if now == nil {
		now = time.Now
	}
	return &Sampler{
		cfg:     cfg,
		buckets: make(map[string]*bucket),
		now:     now,
	}, nil
}

// Allow returns true if the entry identified by key should be forwarded.
// Within each Window, only every (1/Rate)-th entry is passed through.
func (s *Sampler) Allow(key string) bool {
	if s.cfg.Rate >= 1.0 {
		return true
	}
	if s.cfg.Rate <= 0.0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	b, ok := s.buckets[key]
	if !ok || now.After(b.resetAt) {
		b = &bucket{resetAt: now.Add(s.cfg.Window)}
		s.buckets[key] = b
	}
	b.count++

	threshold := int(1.0 / s.cfg.Rate)
	if threshold < 1 {
		threshold = 1
	}
	return b.count%threshold == 1
}

// Reset clears all internal state.
func (s *Sampler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buckets = make(map[string]*bucket)
}
