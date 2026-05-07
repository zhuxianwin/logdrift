package correlation

import (
	"testing"
)

func TestFromConfig_Nil_ReturnsNil(t *testing.T) {
	tr, err := FromConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr != nil {
		t.Fatal("expected nil tracker for nil config")
	}
}

func TestFromConfig_ValidConfig_ReturnsTracker(t *testing.T) {
	cfg := &Config{Field: "trace_id", TTL: "1m"}
	tr, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil tracker")
	}
}

func TestFromConfig_DefaultTTL_UsedWhenEmpty(t *testing.T) {
	cfg := &Config{Field: "request_id", TTL: ""}
	tr, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil tracker")
	}
}

func TestFromConfig_EmptyField_ReturnsError(t *testing.T) {
	cfg := &Config{Field: "", TTL: "1m"}
	_, err := FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for empty field")
	}
}

func TestFromConfig_InvalidTTL_ReturnsError(t *testing.T) {
	cfg := &Config{Field: "trace_id", TTL: "not-a-duration"}
	_, err := FromConfig(cfg)
	if err == nil {
		t.Fatal("expected error for invalid ttl")
	}
}
