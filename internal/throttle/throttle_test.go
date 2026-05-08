package throttle

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_InvalidMaxDiffs(t *testing.T) {
	_, err := New(0, time.Minute)
	if err == nil {
		t.Fatal("expected error for maxDiffs=0")
	}
}

func TestNew_InvalidWindow(t *testing.T) {
	_, err := New(5, 0)
	if err == nil {
		t.Fatal("expected error for window=0")
	}
}

func TestAllow_UnderLimit_ReturnsTrue(t *testing.T) {
	th, _ := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !th.Allow("svc") {
			t.Fatalf("expected Allow=true on call %d", i+1)
		}
	}
}

func TestAllow_OverLimit_ReturnsFalse(t *testing.T) {
	th, _ := New(2, time.Minute)
	th.Allow("svc")
	th.Allow("svc")
	if th.Allow("svc") {
		t.Fatal("expected Allow=false once limit exceeded")
	}
}

func TestAllow_WindowEviction_AllowsAfterExpiry(t *testing.T) {
	base := time.Now()
	th, _ := New(2, time.Second)
	th.now = fixedNow(base)
	th.Allow("svc")
	th.Allow("svc")

	// advance past the window
	th.now = fixedNow(base.Add(2 * time.Second))
	if !th.Allow("svc") {
		t.Fatal("expected Allow=true after window expires")
	}
}

func TestAllow_IndependentServices(t *testing.T) {
	th, _ := New(1, time.Minute)
	th.Allow("a")
	if !th.Allow("b") {
		t.Fatal("service b should not be affected by service a's quota")
	}
}

func TestReset_ClearsService(t *testing.T) {
	th, _ := New(1, time.Minute)
	th.Allow("svc")
	th.Reset("svc")
	if !th.Allow("svc") {
		t.Fatal("expected Allow=true after Reset")
	}
}

func TestFromServiceConfig_Nil_ReturnsNil(t *testing.T) {
	th, err := FromServiceConfig(nil)
	if err != nil || th != nil {
		t.Fatal("expected nil throttle and nil error for nil config")
	}
}

func TestFromServiceConfig_Valid_ReturnsThrottle(t *testing.T) {
	cfg := &ServiceConfig{MaxDiffs: 10, Window: "30s"}
	th, err := FromServiceConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if th == nil {
		t.Fatal("expected non-nil throttle")
	}
}

func TestFromServiceConfig_InvalidWindow_ReturnsError(t *testing.T) {
	cfg := &ServiceConfig{MaxDiffs: 5, Window: "bad"}
	_, err := FromServiceConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid window duration")
	}
}
