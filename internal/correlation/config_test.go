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

func TestFromConfig_MultipleTTLFormats(t *testing.T) {
	tests := []struct {
		ttl   string
		valid bool
	}{
		{"30s", true},
		{"5m", true},
		{"2h", true},
		{"1h30m", true},
		{"0", true},
		{"abc", false},
		{"-1m", false},
	}
	for _, tc := range tests {
		t.Run(tc.ttl, func(t *testing.T) {
			cfg := &Config{Field: "trace_id", TTL: tc.ttl}
			_, err := FromConfig(cfg)
			if tc.valid && err != nil {
				t.Errorf("expected no error for TTL %q, got: %v", tc.ttl, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("expected error for TTL %q, got nil", tc.ttl)
			}
		})
	}
}
