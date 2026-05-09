package ratelimit_test

import (
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/ratelimit"
)

var fixedNow = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNew_InvalidRate_ReturnsError(t *testing.T) {
	_, err := ratelimit.New(-1, 10, func() time.Time { return fixedNow })
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestNew_InvalidBurst_ReturnsError(t *testing.T) {
	_, err := ratelimit.New(5, 0, func() time.Time { return fixedNow })
	if err == nil {
		t.Fatal("expected error for zero burst")
	}
}

func TestAllow_FullBucket_AllowsUpToBurst(t *testing.T) {
	now := fixedNow
	rl, err := ratelimit.New(1, 3, func() time.Time { return now })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 3; i++ {
		if !rl.Allow() {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
	if rl.Allow() {
		t.Fatal("expected deny after burst exhausted")
	}
}

func TestAllow_TokensRefillOverTime(t *testing.T) {
	now := fixedNow
	rl, err := ratelimit.New(1, 1, func() time.Time { return now })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !rl.Allow() {
		t.Fatal("expected first allow")
	}
	if rl.Allow() {
		t.Fatal("expected deny immediately after")
	}

	now = now.Add(2 * time.Second)
	if !rl.Allow() {
		t.Fatal("expected allow after refill")
	}
}

func TestAllow_ZeroRate_BlocksAll(t *testing.T) {
	now := fixedNow
	rl, err := ratelimit.New(0, 1, func() time.Time { return now })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rl.Allow() {
		t.Fatal("expected deny for zero rate")
	}
}

func TestAllow_HighRate_AllowsMany(t *testing.T) {
	now := fixedNow
	rl, err := ratelimit.New(100, 50, func() time.Time { return now })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	allowed := 0
	for i := 0; i < 50; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	if allowed != 50 {
		t.Fatalf("expected 50 allowed, got %d", allowed)
	}
}

func TestAllow_PartialRefill_AllowsCorrectCount(t *testing.T) {
	now := fixedNow
	rl, err := ratelimit.New(2, 2, func() time.Time { return now })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rl.Allow()
	rl.Allow()

	now = now.Add(500 * time.Millisecond)
	if !rl.Allow() {
		t.Fatal("expected one token after 500ms at rate 2/s")
	}
	if rl.Allow() {
		t.Fatal("expected deny after partial refill exhausted")
	}
}
