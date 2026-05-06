package alerting

import (
	"bytes"
	"testing"

	"github.com/user/logdrift/internal/config"
)

func alertConfig(minDelta int, window string) *config.Config {
	return &config.Config{
		Alert: &config.AlertConfig{
			MinDeltaCount: minDelta,
			Window:        window,
		},
	}
}

func TestFromConfig_Nil_ReturnsNil(t *testing.T) {
	cfg := &config.Config{}
	a, err := FromConfig(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != nil {
		t.Fatal("expected nil alerter when no alert config")
	}
}

func TestFromConfig_ValidConfig_ReturnsAlerter(t *testing.T) {
	cfg := alertConfig(2, "5s")
	a, err := FromConfig(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil alerter")
	}
}

func TestFromConfig_InvalidWindow_ReturnsError(t *testing.T) {
	cfg := alertConfig(1, "not-a-duration")
	_, err := FromConfig(cfg, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for invalid window duration")
	}
}

func TestFromConfig_DefaultMinDelta_UsesOne(t *testing.T) {
	cfg := alertConfig(0, "1s")
	a, err := FromConfig(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil alerter")
	}
	if a.threshold.MinDeltaCount != 1 {
		t.Errorf("expected MinDeltaCount=1, got %d", a.threshold.MinDeltaCount)
	}
}

func TestFromConfig_NegativeMinDelta_UsesOne(t *testing.T) {
	cfg := alertConfig(-5, "1s")
	a, err := FromConfig(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil alerter")
	}
	if a.threshold.MinDeltaCount != 1 {
		t.Errorf("expected MinDeltaCount=1 for negative input, got %d", a.threshold.MinDeltaCount)
	}
}
