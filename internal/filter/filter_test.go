package filter_test

import (
	"testing"
	"time"

	"github.com/user/logdrift/internal/filter"
	"github.com/user/logdrift/internal/parser"
)

func makeEntry(fields map[string]any) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test",
		Fields:    fields,
	}
}

func TestNew_InvalidPattern_ReturnsError(t *testing.T) {
	_, err := filter.New([]filter.Rule{{Field: "level", Pattern: "[invalid"}})
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestMatch_AllRulesPass(t *testing.T) {
	f, err := filter.New([]filter.Rule{
		{Field: "level", Pattern: "^error$"},
		{Field: "service", Pattern: "api"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entry := makeEntry(map[string]any{"level": "error", "service": "api-gateway"})
	if !f.Match(entry) {
		t.Error("expected Match to return true")
	}
}

func TestMatch_OneFails_ReturnsFalse(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Pattern: "^error$"},
		{Field: "service", Pattern: "^payments$"},
	})
	entry := makeEntry(map[string]any{"level": "error", "service": "api"})
	if f.Match(entry) {
		t.Error("expected Match to return false when one rule fails")
	}
}

func TestMatch_MissingField_ReturnsFalse(t *testing.T) {
	f, _ := filter.New([]filter.Rule{{Field: "trace_id", Pattern: ".*"}})
	entry := makeEntry(map[string]any{"level": "info"})
	if f.Match(entry) {
		t.Error("expected Match to return false for missing field")
	}
}

func TestMatch_NoRules_ReturnsTrue(t *testing.T) {
	f, _ := filter.New(nil)
	entry := makeEntry(map[string]any{})
	if !f.Match(entry) {
		t.Error("expected Match to return true when no rules defined")
	}
}

func TestMatchAny_OneMatches_ReturnsTrue(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Pattern: "^debug$"},
		{Field: "level", Pattern: "^error$"},
	})
	entry := makeEntry(map[string]any{"level": "error"})
	if !f.MatchAny(entry) {
		t.Error("expected MatchAny to return true")
	}
}

func TestMatchAny_NoneMatch_ReturnsFalse(t *testing.T) {
	f, _ := filter.New([]filter.Rule{
		{Field: "level", Pattern: "^debug$"},
	})
	entry := makeEntry(map[string]any{"level": "info"})
	if f.MatchAny(entry) {
		t.Error("expected MatchAny to return false")
	}
}
