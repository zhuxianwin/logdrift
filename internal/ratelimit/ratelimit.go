// Package ratelimit provides a token-bucket rate limiter for controlling
// how frequently diff output is emitted per service pair.
package ratelimit

import (
	"fmt"
	"sync"
	"time"
)

// Limiter controls emission frequency using a token-bucket algorithm.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	now      func() time.Time
}

// New creates a Limiter that allows up to burst events and refills at
// ratePerSec events per second. Returns an error if either argument is <= 0.
func New(ratePerSec float64, burst float64) (*Limiter, error) {
	if ratePerSec <= 0 {
		return nil, fmt.Errorf("ratePerSec must be > 0, got %v", ratePerSec)
	}
	if burst <= 0 {
		return nil, fmt.Errorf("burst must be > 0, got %v", burst)
	}
	return &Limiter{
		tokens:   burst,
		max:      burst,
		rate:     ratePerSec,
		lastTick: time.Now(),
		now:      time.Now,
	}, nil
}

// Allow returns true if an event is permitted right now, consuming one token.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.lastTick = now

	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}

	if l.tokens >= 1.0 {
		l.tokens--
		return true
	}
	return false
}

// Reset refills the bucket to its maximum capacity.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tokens = l.max
	l.lastTick = l.now()
}
