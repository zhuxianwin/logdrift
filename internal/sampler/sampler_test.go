package sampler

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_InvalidRate(t *testing.T) {
	_, err := New(Config{Rate: 1.5}, nil)
	if err == nil {
		t.Fatal("expected error for rate > 1.0")
	}
	_, err = New(Config{Rate: -0.1}, nil)
	if err == nil {
		t.Fatal("expected error for rate < 0.0")
	}
}

func TestAllow_RateOne_AllowsAll(t *testing.T) {
	s, _ := New(Config{Rate: 1.0, Window: time.Second}, nil)
	for i := 0; i < 20; i++ {
		if !s.Allow("svc") {
			t.Fatalf("expected Allow=true at i=%d", i)
		}
	}
}

func TestAllow_RateZero_BlocksAll(t *testing.T) {
	s, _ := New(Config{Rate: 0.0, Window: time.Second}, nil)
	for i := 0; i < 10; i++ {
		if s.Allow("svc") {
			t.Fatal("expected Allow=false")
		}
	}
}

func TestAllow_RateHalf_AllowsEveryOther(t *testing.T) {
	now := time.Now()
	s, _ := New(Config{Rate: 0.5, Window: time.Minute}, fixedNow(now))

	allowed := 0
	for i := 0; i < 10; i++ {
		if s.Allow("svc") {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("expected 5 allowed out of 10, got %d", allowed)
	}
}

func TestAllow_WindowReset_RestartsCount(t *testing.T) {
	base := time.Now()
	current := base
	nowFn := func() time.Time { return current }

	s, _ := New(Config{Rate: 0.5, Window: time.Second}, nowFn)

	// First window: call twice, only first is allowed.
	s.Allow("svc") // count=1 → allowed
	s.Allow("svc") // count=2 → blocked

	// Advance past window.
	current = base.Add(2 * time.Second)

	// New window: first call should be allowed again.
	if !s.Allow("svc") {
		t.Fatal("expected Allow=true after window reset")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	now := time.Now()
	s, _ := New(Config{Rate: 0.5, Window: time.Minute}, fixedNow(now))

	// Each key has its own counter.
	if !s.Allow("a") {
		t.Fatal("key a count=1 should be allowed")
	}
	if !s.Allow("b") {
		t.Fatal("key b count=1 should be allowed")
	}
	if s.Allow("a") {
		t.Fatal("key a count=2 should be blocked")
	}
	if s.Allow("b") {
		t.Fatal("key b count=2 should be blocked")
	}
}

func TestReset_ClearsState(t *testing.T) {
	now := time.Now()
	s, _ := New(Config{Rate: 0.5, Window: time.Minute}, fixedNow(now))

	s.Allow("svc") // count=1
	s.Allow("svc") // count=2
	s.Reset()

	// After reset, count restarts so first call is allowed.
	if !s.Allow("svc") {
		t.Fatal("expected Allow=true after Reset")
	}
}
