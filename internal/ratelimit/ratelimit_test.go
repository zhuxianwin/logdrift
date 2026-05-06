package ratelimit

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_InvalidRate_ReturnsError(t *testing.T) {
	_, err := New(0, 5)
	if err == nil {
		t.Fatal("expected error for ratePerSec=0")
	}
}

func TestNew_InvalidBurst_ReturnsError(t *testing.T) {
	_, err := New(5, 0)
	if err == nil {
		t.Fatal("expected error for burst=0")
	}
}

func TestAllow_FullBucket_AllowsUpToBurst(t *testing.T) {
	l, _ := New(1, 3)
	base := time.Now()
	l.now = fixedNow(base)
	l.lastTick = base

	allowed := 0
	for i := 0; i < 5; i++ {
		if l.Allow() {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("expected 3 allowed (burst), got %d", allowed)
	}
}

func TestAllow_TokensRefillOverTime(t *testing.T) {
	base := time.Now()
	l, _ := New(2, 2) // 2 tokens/sec, burst 2
	l.now = fixedNow(base)
	l.lastTick = base

	// drain the bucket
	l.Allow()
	l.Allow()

	// advance time by 1 second — should refill 2 tokens
	l.now = fixedNow(base.Add(time.Second))

	if !l.Allow() {
		t.Fatal("expected Allow() after refill")
	}
}

func TestAllow_EmptyBucket_Blocks(t *testing.T) {
	base := time.Now()
	l, _ := New(1, 1)
	l.now = fixedNow(base)
	l.lastTick = base

	l.Allow() // consume the only token

	if l.Allow() {
		t.Fatal("expected Allow() to be blocked on empty bucket")
	}
}

func TestReset_RefillsBucket(t *testing.T) {
	base := time.Now()
	l, _ := New(1, 3)
	l.now = fixedNow(base)
	l.lastTick = base

	l.Allow()
	l.Allow()
	l.Allow()

	l.Reset()

	allowed := 0
	for i := 0; i < 3; i++ {
		if l.Allow() {
			allowed++
		}
	}
	if allowed != 3 {
		t.Fatalf("expected 3 after reset, got %d", allowed)
	}
}

func TestFromServiceConfig_Defaults(t *testing.T) {
	l, err := FromServiceConfig(ServiceConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
}

func TestFromServiceConfig_CustomValues(t *testing.T) {
	l, err := FromServiceConfig(ServiceConfig{RatePerSec: 20, Burst: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l.rate != 20 || l.max != 10 {
		t.Fatalf("expected rate=20 burst=10, got rate=%v burst=%v", l.rate, l.max)
	}
}
