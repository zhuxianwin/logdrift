package alerting

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/logdrift/internal/differ"
)

func makeDelta(field, a, b string) differ.Delta {
	return differ.Delta{Field: field, A: a, B: b}
}

func TestNew_InvalidMinDeltaCount(t *testing.T) {
	_, err := New(Threshold{MinDeltaCount: 0, Window: 0}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for MinDeltaCount=0")
	}
}

func TestNew_NegativeWindow(t *testing.T) {
	_, err := New(Threshold{MinDeltaCount: 1, Window: -time.Second}, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for negative window")
	}
}

func TestEvaluate_BelowThreshold_NoAlert(t *testing.T) {
	buf := &bytes.Buffer{}
	a, _ := New(Threshold{MinDeltaCount: 3, Window: 0}, buf)
	result := a.Evaluate([]differ.Delta{makeDelta("f", "x", "y")})
	if result != nil {
		t.Fatalf("expected nil alert, got %+v", result)
	}
}

func TestEvaluate_WindowNotElapsed_NoAlert(t *testing.T) {
	buf := &bytes.Buffer{}
	base := time.Now()
	a, _ := New(Threshold{MinDeltaCount: 1, Window: 5 * time.Second}, buf)
	a.now = func() time.Time { return base }

	result := a.Evaluate([]differ.Delta{makeDelta("f", "x", "y")})
	if result != nil {
		t.Fatal("expected no alert before window elapses")
	}
}

func TestEvaluate_WindowElapsed_FiresAlert(t *testing.T) {
	buf := &bytes.Buffer{}
	base := time.Now()
	a, _ := New(Threshold{MinDeltaCount: 1, Window: 5 * time.Second}, buf)
	a.now = func() time.Time { return base }
	a.Evaluate([]differ.Delta{makeDelta("f", "x", "y")})

	a.now = func() time.Time { return base.Add(6 * time.Second) }
	result := a.Evaluate([]differ.Delta{makeDelta("f", "x", "y")})
	if result == nil {
		t.Fatal("expected alert after window elapsed")
	}
	if !strings.Contains(buf.String(), "[ALERT]") {
		t.Errorf("expected [ALERT] in output, got: %s", buf.String())
	}
}

func TestEvaluate_BreachClears_ResetsTimer(t *testing.T) {
	buf := &bytes.Buffer{}
	base := time.Now()
	a, _ := New(Threshold{MinDeltaCount: 1, Window: 2 * time.Second}, buf)
	a.now = func() time.Time { return base }
	a.Evaluate([]differ.Delta{makeDelta("f", "x", "y")})

	// clear breach
	a.Evaluate([]differ.Delta{})
	if a.breachSince != nil {
		t.Error("expected breachSince to be reset after clear")
	}
}
