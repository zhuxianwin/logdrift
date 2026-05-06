package filter_test

import (
	"testing"

	"github.com/user/logdrift/internal/filter"
)

func TestFromConfig_EmptySlice_ReturnsAllModeFilter(t *testing.T) {
	f, mode, err := filter.FromConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if mode != "all" {
		t.Errorf("expected mode 'all', got %q", mode)
	}
}

func TestFromConfig_ValidRules_ReturnsFilter(t *testing.T) {
	cfgs := []filter.RuleConfig{
		{Field: "level", Pattern: "error", Mode: "all"},
	}
	f, mode, err := filter.FromConfig(cfgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
	if mode != "all" {
		t.Errorf("expected mode 'all', got %q", mode)
	}
}

func TestFromConfig_AnyMode_Detected(t *testing.T) {
	cfgs := []filter.RuleConfig{
		{Field: "level", Pattern: "error", Mode: "any"},
		{Field: "level", Pattern: "warn"},
	}
	_, mode, err := filter.FromConfig(cfgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "any" {
		t.Errorf("expected mode 'any', got %q", mode)
	}
}

func TestFromConfig_EmptyField_ReturnsError(t *testing.T) {
	cfgs := []filter.RuleConfig{
		{Field: "", Pattern: "error"},
	}
	_, _, err := filter.FromConfig(cfgs)
	if err == nil {
		t.Fatal("expected error for empty field, got nil")
	}
}

func TestFromConfig_EmptyPattern_ReturnsError(t *testing.T) {
	cfgs := []filter.RuleConfig{
		{Field: "level", Pattern: ""},
	}
	_, _, err := filter.FromConfig(cfgs)
	if err == nil {
		t.Fatal("expected error for empty pattern, got nil")
	}
}

func TestFromConfig_InvalidRegex_ReturnsError(t *testing.T) {
	cfgs := []filter.RuleConfig{
		{Field: "level", Pattern: "[bad"},
	}
	_, _, err := filter.FromConfig(cfgs)
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}
