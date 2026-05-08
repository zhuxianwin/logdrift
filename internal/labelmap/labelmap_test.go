package labelmap_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/yourorg/logdrift/internal/labelmap"
	"github.com/yourorg/logdrift/internal/parser"
)

func makeEntry(fields map[string]interface{}) parser.Entry {
	return parser.Entry{
		Service:   "svc",
		Timestamp: time.Now(),
		Raw:       `{}`,
		Fields:    fields,
	}
}

func TestNew_EmptySourceField_ReturnsError(t *testing.T) {
	_, err := labelmap.New([]labelmap.Rule{
		{SourceField: "", DestField: "env", Pattern: regexp.MustCompile(`.*`), Label: "prod"},
	})
	if err == nil {
		t.Fatal("expected error for empty source_field")
	}
}

func TestNew_EmptyDestField_ReturnsError(t *testing.T) {
	_, err := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "", Pattern: regexp.MustCompile(`.*`), Label: "prod"},
	})
	if err == nil {
		t.Fatal("expected error for empty dest_field")
	}
}

func TestNew_NilPattern_ReturnsError(t *testing.T) {
	_, err := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "env", Pattern: nil, Label: "prod"},
	})
	if err == nil {
		t.Fatal("expected error for nil pattern")
	}
}

func TestNew_EmptyLabel_ReturnsError(t *testing.T) {
	_, err := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "env", Pattern: regexp.MustCompile(`.*`), Label: ""},
	})
	if err == nil {
		t.Fatal("expected error for empty label")
	}
}

func TestApply_NoRules_ReturnsEntryUnchanged(t *testing.T) {
	m, err := labelmap.New(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := makeEntry(map[string]interface{}{"level": "error"})
	out := m.Apply(e)
	if out.Fields["level"] != "error" {
		t.Errorf("expected level=error, got %v", out.Fields["level"])
	}
}

func TestApply_MatchingRule_InjectsLabel(t *testing.T) {
	m, err := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "severity_label", Pattern: regexp.MustCompile(`^error$`), Label: "high"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := makeEntry(map[string]interface{}{"level": "error"})
	out := m.Apply(e)
	if out.Fields["severity_label"] != "high" {
		t.Errorf("expected severity_label=high, got %v", out.Fields["severity_label"])
	}
}

func TestApply_NonMatchingRule_DoesNotInjectLabel(t *testing.T) {
	m, err := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "severity_label", Pattern: regexp.MustCompile(`^error$`), Label: "high"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	e := makeEntry(map[string]interface{}{"level": "info"})
	out := m.Apply(e)
	if _, ok := out.Fields["severity_label"]; ok {
		t.Error("expected severity_label to be absent")
	}
}

func TestApply_DoesNotMutateOriginalEntry(t *testing.T) {
	m, _ := labelmap.New([]labelmap.Rule{
		{SourceField: "level", DestField: "tag", Pattern: regexp.MustCompile(`.*`), Label: "tagged"},
	})
	origFields := map[string]interface{}{"level": "warn"}
	e := makeEntry(origFields)
	m.Apply(e)
	if _, ok := origFields["tag"]; ok {
		t.Error("original fields map was mutated")
	}
}
